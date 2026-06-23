package nats

import (
	"context"
	"encoding/json"
	"fmt"

	"internal/application/port"
	"internal/domain"

	"github.com/nats-io/nats.go"
)

var _ port.EventPublisher = (*EventPublisher)(nil)

const (
	SubjectUserCreated = "user.created"
	SubjectOTPCreated  = "otp.created"
)

type EventPublisher struct {
	js nats.JetStreamContext
}

func (e *EventPublisher) PublishUserCreated(ctx context.Context, evt domain.UserCreated) error {
	return e.publish(ctx, SubjectUserCreated, "UserCreated", evt)
}

func (e *EventPublisher) PublishOTPCreated(ctx context.Context, evt domain.OTPCreated) error {
	return e.publish(ctx, SubjectOTPCreated, "OTPCreated", evt)
}

func (e *EventPublisher) publish(ctx context.Context, subject, eventType string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("nats - failed to marshal %s: %w", eventType, err)
	}

	msg := &nats.Msg{
		Subject: subject,
		Data:    data,
		Header:  nats.Header{},
	}
	msg.Header.Set("Event-Type", eventType)

	_, err = e.js.PublishMsg(msg, nats.Context(ctx))
	if err != nil {
		return fmt.Errorf("nats - failed to publish %s to %q: %w", eventType, subject, err)
	}

	return nil
}
