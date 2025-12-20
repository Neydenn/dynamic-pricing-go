package services

import "context"

// EventBus abstracts message publishing (e.g., Kafka producer).
type EventBus interface {
    Send(ctx context.Context, key string, value []byte) error
}

