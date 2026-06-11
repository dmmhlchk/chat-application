package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"identity-service/internal/handler/http/middleware"
	v1 "identity-service/internal/handler/http/v1"
	"identity-service/internal/infrastructure/cache"
	"identity-service/internal/infrastructure/crypto"
	"identity-service/internal/infrastructure/db/postgres"
	"identity-service/internal/infrastructure/sms"
	"identity-service/internal/usecase"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

func main() {
	// ----------------------------------------------------------------------
	// 				GETTING CONFIGS
	// ----------------------------------------------------------------------

	postgresURL := getEnvOrPanic("POSTGRES_URL")
	redisURL := getEnvOrPanic("REDIS_URL")
	jwtSecret := getEnvOrPanic("JWT_SECRET")

	// ----------------------------------------------------------------------
	//  			INFRASTRUCTURE: DB & CACHE INITIALIZATION
	// ----------------------------------------------------------------------

	pdb := initPostgres(postgresURL)
	defer pdb.Close()
	userRepo := postgres.NewUserRepository(pdb)
	sessionRepo := postgres.NewSessionRepository(pdb)

	rdb := initRedis(redisURL)
	defer rdb.Close()
	otpRepo := cache.NewRedisOTPGenerator(rdb, 2*time.Minute)

	// ----------------------------------------------------------------------
	//  			INFRASTRUCTURE: CRYPTO & SMS PROVIDER INITIALIZATION
	// ----------------------------------------------------------------------

	jwtGen := crypto.NewJWTGenerator(jwtSecret)
	otpGen := crypto.NewSecureOTPGenerator()
	pwdHasher := crypto.NewBcryptHasher()
	dummySMS := sms.NewDummySMSProvider()

	// ----------------------------------------------------------------------
	//  			USE CASE INITIALIZATION
	// ----------------------------------------------------------------------

	userSignUpReqUC := usecase.NewSignUpRequest(userRepo, dummySMS, otpGen, otpRepo)
	userSignUpConfUC := usecase.NewSignUpConfirm(userRepo, otpRepo, pwdHasher)
	userSignInUC := usecase.NewSignIn(userRepo, sessionRepo, pwdHasher, jwtGen)
	userDeletionUC := usecase.NewUserDelete(userRepo, sessionRepo, pwdHasher)
	sessionTerminationUC := usecase.NewTerminateSession(sessionRepo, jwtGen)
	sessioListUC := usecase.NewSessionList(sessionRepo)
	passwordChangeUC := usecase.NewChangePassword(userRepo, pwdHasher)
	passwordResetReqUC := usecase.NewPasswordResetRequest(userRepo, dummySMS, otpGen, otpRepo)
	passwordResetConfUC := usecase.NewPasswordResetConfirm(userRepo, sessionRepo, otpRepo, pwdHasher)

	// ----------------------------------------------------------------------
	//  			DELIVERY LAYER INITIALIZATION
	// ----------------------------------------------------------------------

	userSignUp := v1.NewSignUpHandler(userSignUpReqUC, userSignUpConfUC)
	userSignIn := v1.NewSignInHandler(userSignInUC)
	userDeletion := v1.NewUserDeletion(userDeletionUC)
	sessionTermination := v1.NewTerminateSession(sessionTerminationUC)
	sessioList := v1.NewSessionList(sessioListUC)
	passwordChange := v1.NewChangePassword(passwordChangeUC)
	passwordReset := v1.NewPasswordReset(passwordResetReqUC, passwordResetConfUC)

	// ----------------------------------------------------------------------
	//  			HTTP ROUTER & SERVER INITIALIZATION
	// ----------------------------------------------------------------------

	mux := http.NewServeMux()

	// Public Routes (No authentication required)
	mux.HandleFunc("POST /api/v1/auth/signup/request", userSignUp.HandleRequest)
	mux.HandleFunc("POST /api/v1/auth/signup/confirm", userSignUp.HandleConfirm)
	mux.HandleFunc("POST /api/v1/auth/signin", userSignIn.Handle)
	mux.HandleFunc("POST /api/v1/auth/password/reset/request", passwordReset.HandleRequest)
	mux.HandleFunc("POST /api/v1/auth/password/reset/confirm", passwordReset.HandleConfirm)

	// Protected Routes (Technically need Auth, but passing raw requests for now)
	mux.HandleFunc("GET /api/v1/sessions", sessioList.Handle)
	mux.HandleFunc("POST /api/v1/sessions/", sessionTermination.Handle) // e.g., /api/v1/sessions/{id}
	mux.HandleFunc("POST /api/v1/auth/password/change", passwordChange.Handle)
	mux.HandleFunc("POST /api/v1/users/me", userDeletion.Handle)

	loggedRouter := middleware.RequestLogger(mux)

	// Define Server Parameters
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      loggedRouter,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("Identity Service API listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server stopped unexpectedly: %v", err)
		}
	}()

	// ----------------------------------------------------------------------
	//              GRACEFUL SHUTDOWN TRAP
	// ----------------------------------------------------------------------

	// Listen for interrupt signals (Ctrl+C or Docker stop requests)
	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, os.Interrupt, syscall.SIGTERM)

	<-stopSignal // Block here until a signal is received
	log.Println("Shutdown signal intercepted. Clearing application tasks...")

	// Give active network requests 15 seconds to naturally finish processing
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to crash on shutdown: %v", err)
	}

	log.Println("Identity Service safely offline.")
}

func getEnvOrPanic(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("Boot error: Environment variable %q required.\n", key)
		os.Exit(1)
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
