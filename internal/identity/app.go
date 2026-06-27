package identity

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"chat-app/internal/identity/application/usecase"
	"chat-app/internal/identity/infra/identifier"
	"chat-app/internal/identity/infra/messaging/nats"
	"chat-app/internal/identity/infra/persistence/postgres"
	reporedis "chat-app/internal/identity/infra/persistence/redis"
	"chat-app/internal/identity/infra/security/otp"
	"chat-app/internal/identity/infra/security/password"
	"chat-app/internal/identity/infra/security/token"

	v1 "chat-app/internal/identity/api/http/v1"
	"chat-app/internal/middleware"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

// Container holds all instantiated components for the application lifecycle
type Container struct {
	PostgresDB *sql.DB
	RedisDB    *redis.Client
	NatsClient *nats.Client
	Router     http.Handler
}

// NewContainer initializes configurations, infrastructure, use cases, and routes.
func NewContainer() *Container {
	// 1. GETTING CONFIGS
	postgresURL := getEnvOrPanic("POSTGRES_URL")
	redisURL := getEnvOrPanic("REDIS_URL")
	natsURL := getEnvOrPanic("NATS_URL")
	jwtSecret := getEnvOrPanic("JWT_SECRET")

	// 2. INFRASTRUCTURE: DB & CACHE INITIALIZATION
	postgresdb := initPostgres(postgresURL)
	userRepo := postgres.NewUserRepository(postgresdb)
	sessionRepo := postgres.NewSessionRepository(postgresdb)

	redisdb := initRedis(redisURL)
	otpRepo := reporedis.NewOTPCacheRepository(redisdb)

	// 3. INFRASTRUCTURE: NATS JETSTREAM INITIALIZATION
	natsClient := initNats(natsURL)
	jsCtx := *natsClient.JetStream()

	// Ensure your identity stream configuration exists on startup
	identityStreamCfg := nats.DefaultStreamConfig(
		"IDENTITY_EVENTS",
		[]string{nats.SubjectUserCreated, nats.SubjectOTPCreated},
	)
	if err := nats.EnsureStream(jsCtx, identityStreamCfg); err != nil {
		log.Fatalf("Critical error establishing NATS JetStream topology: %v", err)
	}

	eventPublisher := nats.NewEventPublisher(jsCtx)

	// 4. INFRASTRUCTURE: IDENTIFIER & SECURITY INITIALIZATION
	uuidGen := identifier.NewUUIDGenerator()
	jwtGen := token.NewJWTGenerator(jwtSecret)
	otpGen := otp.NewSecureOTPGenerator()
	passwordHasher := password.NewBcryptPasswordHasher()

	// 5. APPLICATION: USE CASE INITIALIZATION
	userSignUpReqUC := usecase.NewSignUpRequest(userRepo, eventPublisher, otpGen, otpRepo)
	userSignUpConfUC := usecase.NewSignUpConfirm(uuidGen, userRepo, otpRepo, passwordHasher)
	userSignInUC := usecase.NewSignIn(uuidGen, userRepo, sessionRepo, passwordHasher, jwtGen)
	userDeletionUC := usecase.NewUserDelete(userRepo, passwordHasher)
	sessionTerminationUC := usecase.NewTerminateSession(sessionRepo, jwtGen)
	sessioListUC := usecase.NewSessionList(sessionRepo)
	passwordChangeUC := usecase.NewChangePassword(userRepo, sessionRepo, passwordHasher)
	passwordResetReqUC := usecase.NewPasswordResetRequest(userRepo, eventPublisher, otpGen, otpRepo)
	passwordResetConfUC := usecase.NewPasswordResetConfirm(userRepo, sessionRepo, otpRepo, passwordHasher)

	// 6. DELIVERY LAYER INITIALIZATION
	userSignUpReq := v1.NewSignUpRequestHandler(userSignUpReqUC)
	userSignUpConf := v1.NewSignUpConfirmHandler(userSignUpConfUC)
	userSignIn := v1.NewSignInHandler(userSignInUC)
	userDeletion := v1.NewUserDeletion(userDeletionUC)
	sessionTermination := v1.NewTerminateSession(sessionTerminationUC)
	sessioList := v1.NewSessionList(sessioListUC)
	passwordChange := v1.NewChangePassword(passwordChangeUC)
	passwordResetReq := v1.NewPasswordResetRequest(passwordResetReqUC)
	passwordResetConf := v1.NewPasswordResetConfirm(passwordResetConfUC)

	// 7. HTTP ROUTER INITIALIZATION
	mux := http.NewServeMux()

	// Public Routes
	mux.HandleFunc("POST /api/v1/auth/signup/request", userSignUpReq.Handle)
	mux.HandleFunc("POST /api/v1/auth/signup/confirm", userSignUpConf.Handle)
	mux.HandleFunc("POST /api/v1/auth/signin", userSignIn.Handle)
	mux.HandleFunc("POST /api/v1/auth/password/reset/request", passwordResetReq.Handle)
	mux.HandleFunc("POST /api/v1/auth/password/reset/confirm", passwordResetConf.Handle)

	// Protected Routes
	mux.HandleFunc("GET /api/v1/sessions", sessioList.Handle)
	mux.HandleFunc("POST /api/v1/sessions/", sessionTermination.Handle)
	mux.HandleFunc("POST /api/v1/auth/password/change", passwordChange.Handle)
	mux.HandleFunc("POST /api/v1/users/me", userDeletion.Handle)

	loggedRouter := middleware.RequestLogger(mux)

	return &Container{
		PostgresDB: postgresdb,
		RedisDB:    redisdb,
		NatsClient: natsClient,
		Router:     loggedRouter,
	}
}

// Close handles safe cleanup of all infrastructure resource pools
func (c *Container) Close() {
	if c.PostgresDB != nil {
		log.Println("Closing Postgres connection pool...")
		c.PostgresDB.Close()
	}
	if c.RedisDB != nil {
		log.Println("Closing Redis client...")
		c.RedisDB.Close()
	}
	if c.NatsClient != nil {
		log.Println("Draining and closing NATS client connections...")
		c.NatsClient.Close()
	}
}

// --- Helper Initialization Functions ---

func getEnvOrPanic(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("Boot error: Environment variable %q required.\n", key)
	}
	return val
}

func initPostgres(connString string) *sql.DB {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		log.Fatalf("Critical error defining Postgres connection pool layout: %v", err)
	}

	db.SetMaxOpenConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetMaxIdleConns(25)
	db.SetConnMaxIdleTime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Critical error verifying database connection: %v", err)
	}
	return db
}

func initRedis(connString string) *redis.Client {
	opts, err := redis.ParseURL(connString)
	if err != nil {
		log.Fatalf("Critical error parsing Redis URL layout: %v", err)
	}

	opts.PoolSize = 20
	opts.MinIdleConns = 5

	rdb := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Critical error verifying Redis connection: %v", err)
	}
	return rdb
}

func initNats(url string) *nats.Client {
	cfg := nats.Config{
		URL:            url,
		MaxReconnects:  5,
		ReconnectWait:  2 * time.Second,
		ConnectTimeout: 5 * time.Second,
	}

	client, err := nats.NewClient(cfg)
	if err != nil {
		log.Fatalf("Critical error initializing NATS JetStream client: %v", err)
	}
	return client
}
