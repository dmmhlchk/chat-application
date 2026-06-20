package nats

import (
	"errors"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

type StreamConfig struct {
	Name      string
	Subjects  []string
	MaxAge    time.Duration // how long to retain messages; 0 = forever
	Storage   nats.StorageType
	Retention nats.RetentionPolicy
}

func DefaultStreamConfig(name string, subjects []string) StreamConfig {
	return StreamConfig{
		Name:      name,
		Subjects:  subjects,
		MaxAge:    7 * 24 * time.Hour,
		Storage:   nats.FileStorage,
		Retention: nats.LimitsPolicy,
	}
}

// EnsureStream creates the stream if it doesn't exist, or verifies/updates
// it if it does. This is what makes it safe to call on every app startup.
func EnsureStream(js nats.JetStreamContext, cfg StreamConfig) error {
	existing, err := js.StreamInfo(cfg.Name)

	if err != nil {
		// ErrStreamNotFound is the *expected* error on first run - not a
		// real failure. Any other error means something is actually wrong
		// (e.g., can't reach NATS), so we propagate it.
		if !errors.Is(err, nats.ErrStreamNotFound) {
			return fmt.Errorf("nats: failed to check stream %q: %w", cfg.Name, err)
		}

		_, err = js.AddStream(&nats.StreamConfig{
			Name:      cfg.Name,
			Subjects:  cfg.Subjects,
			MaxAge:    cfg.MaxAge,
			Storage:   cfg.Storage,
			Retention: cfg.Retention,
		})
		if err != nil {
			return fmt.Errorf("nats: failed to create stream %q: %w", cfg.Name, err)
		}
		return nil
	}

	// Stream already exists. We don't blindly overwrite it - subject list
	// changes are the most common reason you'd need an update, so that's
	// the one case we handle automatically. Other changes (storage type,
	// retention policy) are left alone since NATS restricts or disallows
	// changing them on a live stream anyway.
	if !subjectsEqual(existing.Config.Subjects, cfg.Subjects) {
		_, err = js.UpdateStream(&nats.StreamConfig{
			Name:      cfg.Name,
			Subjects:  cfg.Subjects,
			MaxAge:    cfg.MaxAge,
			Storage:   cfg.Storage,
			Retention: cfg.Retention,
		})
		if err != nil {
			return fmt.Errorf("nats: failed to update stream %q: %w", cfg.Name, err)
		}
	}

	return nil
}

func subjectsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	seen := make(map[string]bool, len(a))
	for _, s := range a {
		seen[s] = true
	}
	for _, s := range b {
		if !seen[s] {
			return false
		}
	}
	return true
}
