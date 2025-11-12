package sns

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

// SNSAPI defines the subset of sns.Client methods we use.
// This makes it mockable in tests.
type SNSAPI interface {
	Publish(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error)
}

// Publisher defines the interface for publishing SNS messages.
type Publisher interface {
	PublishJSON(ctx context.Context, topicARN string, payload any) (string, error)
	PublishString(ctx context.Context, topicARN, message string) (string, error)
}

// Client wraps an AWS SNS client with helpers.
type Client struct {
	snsClient  SNSAPI
	defaultARN string
}

// Config holds optional configuration for SNS setup.
type Config struct {
	Region   string
	TopicARN string
}

// NewConfigFromEnv builds configuration from environment variables.
func NewConfigFromEnv() (*Config, error) {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		return nil, fmt.Errorf("AWS_REGION is required")
	}

	topic := os.Getenv("SNS_TOPIC_ARN")
	return &Config{
		Region:   region,
		TopicARN: topic,
	}, nil
}

// New creates a new SNS client from AWS credentials/config in the environment.
func New(cfg *Config) (*Client, error) {
	awsCfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(cfg.Region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS configuration: %w", err)
	}
	return &Client{
		snsClient:  sns.NewFromConfig(awsCfg),
		defaultARN: cfg.TopicARN,
	}, nil
}

// PublishString publishes a plain string message to an SNS topic.
func (c *Client) PublishString(ctx context.Context, topicARN, message string) (string, error) {
	if topicARN == "" {
		topicARN = c.defaultARN
	}
	if topicARN == "" {
		return "", fmt.Errorf("topic ARN is required")
	}

	out, err := c.snsClient.Publish(ctx, &sns.PublishInput{
		Message:  aws.String(message),
		TopicArn: aws.String(topicARN),
	})
	if err != nil {
		return "", fmt.Errorf("publish failed: %w", err)
	}
	return aws.ToString(out.MessageId), nil
}

// PublishJSON marshals a struct as JSON and publishes it.
func (c *Client) PublishJSON(ctx context.Context, topicARN string, payload any) (string, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return c.PublishString(ctx, topicARN, string(data))
}
