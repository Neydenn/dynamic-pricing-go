package services

import "context"

type EventBus interface {
	Send(ctx context.Context, key string, value []byte) error
}
