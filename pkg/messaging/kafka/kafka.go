// pkg/messaging/kafka/kafka.go
package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/IBM/sarama"
)

// Config defines Kafka connection and client options.
type Config struct {
	Brokers  []string
	ClientID string
	Version  string
}

// Producer wraps a Sarama async producer for publishing messages.
type Producer struct {
	client   sarama.SyncProducer
	producer sarama.SyncProducer
}

// Consumer wraps a Sarama consumer group for message processing.
type Consumer struct {
	group   sarama.ConsumerGroup
	topics  []string
	handler MessageHandler
}

// MessageHandler defines the signature for handling consumed messages.
type MessageHandler interface {
	HandleMessage(ctx context.Context, msg *sarama.ConsumerMessage) error
}

// NewConfigFromEnv loads Kafka configuration from environment variables.
func NewConfigFromEnv() (*Config, error) {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		return nil, fmt.Errorf("KAFKA_BROKERS is required")
	}
	clientID := os.Getenv("KAFKA_CLIENT_ID")
	if clientID == "" {
		clientID = "http-common-go"
	}
	version := os.Getenv("KAFKA_VERSION")
	if version == "" {
		version = "2.8.0"
	}
	return &Config{
		Brokers:  []string{brokers},
		ClientID: clientID,
		Version:  version,
	}, nil
}

// NewProducer initializes a new Kafka SyncProducer.
func NewProducer(cfg *Config) (*Producer, error) {
	version, err := sarama.ParseKafkaVersion(cfg.Version)
	if err != nil {
		return nil, fmt.Errorf("invalid Kafka version: %w", err)
	}

	saramaCfg := sarama.NewConfig()
	saramaCfg.Producer.Return.Successes = true
	saramaCfg.Producer.RequiredAcks = sarama.WaitForAll
	saramaCfg.Producer.Retry.Max = 5
	saramaCfg.ClientID = cfg.ClientID
	saramaCfg.Version = version

	prod, err := sarama.NewSyncProducer(cfg.Brokers, saramaCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	return &Producer{client: prod, producer: prod}, nil
}

// SendJSON publishes a JSON-encoded message to a Kafka topic.
func (p *Producer) SendJSON(ctx context.Context, topic string, key string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(data),
	}

	_, _, err = p.producer.SendMessage(msg)
	return err
}

// Close shuts down the producer.
func (p *Producer) Close() error {
	if p.producer != nil {
		return p.producer.Close()
	}
	return nil
}

// NewConsumer creates a new Kafka consumer group.
func NewConsumer(cfg *Config, groupID string, topics []string, handler MessageHandler) (*Consumer, error) {
	version, err := sarama.ParseKafkaVersion(cfg.Version)
	if err != nil {
		return nil, fmt.Errorf("invalid Kafka version: %w", err)
	}

	saramaCfg := sarama.NewConfig()
	saramaCfg.Consumer.Offsets.Initial = sarama.OffsetNewest
	saramaCfg.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	saramaCfg.Version = version
	saramaCfg.ClientID = cfg.ClientID

	group, err := sarama.NewConsumerGroup(cfg.Brokers, groupID, saramaCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer group: %w", err)
	}

	return &Consumer{
		group:   group,
		topics:  topics,
		handler: handler,
	}, nil
}

// Run starts consuming messages from configured topics until context is canceled.
func (c *Consumer) Run(ctx context.Context) error {
	for {
		if err := c.group.Consume(ctx, c.topics, &consumerGroupHandler{handler: c.handler}); err != nil {
			return fmt.Errorf("consume error: %w", err)
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

// Close closes the consumer group.
func (c *Consumer) Close() error {
	return c.group.Close()
}

// consumerGroupHandler bridges Sarama's interface to our MessageHandler.
type consumerGroupHandler struct {
	handler MessageHandler
}

func (h *consumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h *consumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		ctx := context.Background()
		if err := h.handler.HandleMessage(ctx, msg); err == nil {
			sess.MarkMessage(msg, "")
		}
	}
	return nil
}
