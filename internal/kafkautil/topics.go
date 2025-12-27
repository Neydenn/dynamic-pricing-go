package kafkautil

import (
    "context"
    "fmt"
    "log/slog"
    "net"
    "net/netip"
    "time"

    "github.com/segmentio/kafka-go"
)

// EnsureTopic creates the topic if it doesn't exist. Idempotent on existing topics.
func EnsureTopic(parent context.Context, brokers []string, topic string, partitions int, replicationFactor int) error {
    if topic == "" {
        return fmt.Errorf("empty topic")
    }
    if len(brokers) == 0 {
        return fmt.Errorf("no kafka brokers configured")
    }

    ctx, cancel := context.WithTimeout(parent, 10*time.Second)
    defer cancel()

    // Dial any broker to find the controller.
    dialer := &kafka.Dialer{Timeout: 5 * time.Second}
    conn, err := dialer.DialContext(ctx, "tcp", brokers[0])
    if err != nil {
        return fmt.Errorf("dial broker: %w", err)
    }
    defer conn.Close()

    ctrl, err := conn.Controller()
    if err != nil {
        return fmt.Errorf("get controller: %w", err)
    }

    // Build controller network address.
    host := ctrl.Host
    port := ctrl.Port
    // Normalize IPv6 host if needed.
    if ip, err := netip.ParseAddr(host); err == nil && ip.Is6() {
        host = "[" + host + "]"
    }
    addr := net.JoinHostPort(host, fmt.Sprint(port))

    ctrlConn, err := dialer.DialContext(ctx, "tcp", addr)
    if err != nil {
        return fmt.Errorf("dial controller: %w", err)
    }
    defer ctrlConn.Close()

    // Create topic (idempotent; TopicAlreadyExists is ignored inside kafka-go).
    if err := ctrlConn.CreateTopics(kafka.TopicConfig{
        Topic:             topic,
        NumPartitions:     partitions,
        ReplicationFactor: replicationFactor,
    }); err != nil {
        return fmt.Errorf("create topic %s: %w", topic, err)
    }
    slog.Info("kafka: ensured topic", "topic", topic)
    return nil
}

