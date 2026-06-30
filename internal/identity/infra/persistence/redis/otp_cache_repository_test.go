package redis_test

import (
	"context"
	"testing"
	"time"

	"chat-app/internal/identity/application/repository"
	"chat-app/internal/identity/domain"
	reporedis "chat-app/internal/identity/infra/persistence/redis"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// ___ Helpers _________________________________________________________________
func newTestRepo(t *testing.T) (repository.OTPCacheRepository, *miniredis.Miniredis) {
	t.Helper()

	mr := miniredis.RunT(t) // automatically stopped when the test ends

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	repo := reporedis.NewOTPCacheRepository(client)
	return repo, mr
}

// ___ Tests _________________________________________________________________
func TestSave_StoresKeyWithTTL(t *testing.T) {
	repo, mr := newTestRepo(t)
	ctx := context.Background()

	ttl := 2 * time.Minute
	err := repo.Save(ctx, "+996700000001", "123456", ttl)
	if err != nil {
		t.Fatalf("Save() unexpected error: %v", err)
	}

	// Verify the key exists in the underlying mini-redis store.
	if !mr.Exists("otp:+996700000001") {
		t.Fatal("expected key 'otp:+996700000001' to exist in Redis")
	}

	// Verify the stored value.
	val, err := mr.Get("otp:+996700000001")
	if err != nil {
		t.Fatalf("miniredis Get: %v", err)
	}
	if val != "123456" {
		t.Errorf("stored value = %q, want %q", val, "123456")
	}

	// Verify the TTL is set (miniredis rounds to seconds).
	gotTTL := mr.TTL("otp:+996700000001")
	if gotTTL <= 0 {
		t.Errorf("expected a positive TTL, got %v", gotTTL)
	}
}

func TestSave_OverwritesExistingCode(t *testing.T) {
	repo, mr := newTestRepo(t)
	ctx := context.Background()

	_ = repo.Save(ctx, "+996700000001", "111111", time.Minute)
	err := repo.Save(ctx, "+996700000001", "999999", time.Minute)
	if err != nil {
		t.Fatalf("second Save() unexpected error: %v", err)
	}

	val, _ := mr.Get("otp:+996700000001")
	if val != "999999" {
		t.Errorf("expected overwritten value %q, got %q", "999999", val)
	}
}

func TestVerify_ReturnsTrueForCorrectCode(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()

	_ = repo.Save(ctx, "+996700000002", "654321", time.Minute)

	ok, err := repo.Verify(ctx, "+996700000002", "654321")
	if err != nil {
		t.Fatalf("Verify() unexpected error: %v", err)
	}
	if !ok {
		t.Error("Verify() = false, want true")
	}
}

func TestVerify_ReturnsFalseAndErrOTPInvalidForWrongCode(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()

	_ = repo.Save(ctx, "+996700000003", "111111", time.Minute)

	ok, err := repo.Verify(ctx, "+996700000003", "000000")
	if ok {
		t.Error("Verify() = true, want false")
	}
	if err != domain.ErrOTPInvalid {
		t.Errorf("Verify() error = %v, want domain.ErrOTPInvalid", err)
	}
}

func TestVerify_ReturnsFalseAndErrOTPExpiredWhenKeyMissing(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()

	// No Save — key simply doesn't exist.
	ok, err := repo.Verify(ctx, "+996700000099", "123456")
	if ok {
		t.Error("Verify() = true, want false")
	}
	if err != domain.ErrOTPExpired {
		t.Errorf("Verify() error = %v, want domain.ErrOTPExpired", err)
	}
}

func TestVerify_ReturnsFalseAndErrOTPExpiredAfterTTLElapses(t *testing.T) {
	repo, mr := newTestRepo(t)
	ctx := context.Background()

	_ = repo.Save(ctx, "+996700000004", "777777", 1*time.Second)

	// Fast-forward miniredis clock so the key expires.
	mr.FastForward(2 * time.Second)

	ok, err := repo.Verify(ctx, "+996700000004", "777777")
	if ok {
		t.Error("Verify() = true after TTL, want false")
	}
	if err != domain.ErrOTPExpired {
		t.Errorf("Verify() error = %v, want domain.ErrOTPExpired", err)
	}
}

func TestDelete_RemovesKey(t *testing.T) {
	repo, mr := newTestRepo(t)
	ctx := context.Background()

	_ = repo.Save(ctx, "+996700000005", "123456", time.Minute)
	err := repo.Delete(ctx, "+996700000005")
	if err != nil {
		t.Fatalf("Delete() unexpected error: %v", err)
	}

	if mr.Exists("otp:+996700000005") {
		t.Error("expected key to be deleted, but it still exists")
	}
}

func TestDelete_IsIdempotentForMissingKey(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()

	// Deleting a key that was never saved must not return an error.
	err := repo.Delete(ctx, "+996700000099")
	if err != nil {
		t.Errorf("Delete() on missing key returned error: %v", err)
	}
}

func TestKeyIsolation_DifferentPhonesAreIndependent(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()

	_ = repo.Save(ctx, "+996700000010", "AAAA", time.Minute)
	_ = repo.Save(ctx, "+996700000011", "BBBB", time.Minute)

	ok, err := repo.Verify(ctx, "+996700000010", "BBBB")
	if ok || err != domain.ErrOTPInvalid {
		t.Errorf("phone A should not verify with phone B's code: ok=%v err=%v", ok, err)
	}

	ok, err = repo.Verify(ctx, "+996700000011", "AAAA")
	if ok || err != domain.ErrOTPInvalid {
		t.Errorf("phone B should not verify with phone A's code: ok=%v err=%v", ok, err)
	}
}
