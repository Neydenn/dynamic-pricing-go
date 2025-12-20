package consumer

import (
    "context"

    "github.com/segmentio/kafka-go"
)

type Consumer struct { r *kafka.Reader }

func New(brokers []string, topic string, groupID string) *Consumer {
    return &Consumer{ r: kafka.NewReader(kafka.ReaderConfig{
        Brokers: brokers,
        Topic:   topic,
        GroupID: groupID,
    })}
}

func (c *Consumer) Close() error { return c.r.Close() }

func (c *Consumer) Read(ctx context.Context) (kafka.Message, error) {
    return c.r.ReadMessage(ctx)
}

