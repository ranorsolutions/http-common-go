package sns

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/stretchr/testify/assert"
)

// mockSNSClient fakes SNSAPI for testing.
type mockSNSClient struct {
	lastInput *sns.PublishInput
	err       error
}

func (m *mockSNSClient) Publish(ctx context.Context, input *sns.PublishInput, _ ...func(*sns.Options)) (*sns.PublishOutput, error) {
	m.lastInput = input
	if m.err != nil {
		return nil, m.err
	}
	return &sns.PublishOutput{MessageId: aws.String("msg-123")}, nil
}

func TestPublishString_Success(t *testing.T) {
	mock := &mockSNSClient{}
	c := &Client{snsClient: mock, defaultARN: "arn:aws:sns:us-east-1:123456789012:test"}

	id, err := c.PublishString(context.Background(), "", "hello world")

	assert.NoError(t, err)
	assert.Equal(t, "msg-123", id)
}

func TestPublishString_Error(t *testing.T) {
	mock := &mockSNSClient{err: errors.New("failed")}
	c := &Client{snsClient: mock, defaultARN: "arn:aws:sns:us-east-1:123456789012:test"}

	_, err := c.PublishString(context.Background(), "", "hi")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed")
}

func TestPublishJSON_Success(t *testing.T) {
	mock := &mockSNSClient{}
	c := &Client{snsClient: mock, defaultARN: "arn:aws:sns:us-east-1:123456789012:test"}

	payload := map[string]string{"event": "user.created"}
	id, err := c.PublishJSON(context.Background(), "", payload)

	assert.NoError(t, err)
	assert.Equal(t, "msg-123", id)

	var m map[string]string
	assert.NoError(t, json.Unmarshal([]byte(aws.ToString(mock.lastInput.Message)), &m))
	assert.Equal(t, "user.created", m["event"])
}

func TestPublishJSON_MarshalError(t *testing.T) {
	mock := &mockSNSClient{}
	c := &Client{snsClient: mock, defaultARN: "arn:aws:sns:us-east-1:123456789012:test"}

	ch := make(chan int)
	defer close(ch)

	_, err := c.PublishJSON(context.Background(), "", ch)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to marshal JSON")
}

func TestNewConfigFromEnv(t *testing.T) {
	t.Setenv("AWS_REGION", "us-west-2")
	t.Setenv("SNS_TOPIC_ARN", "arn:aws:sns:us-west-2:123:test-topic")

	cfg, err := NewConfigFromEnv()
	assert.NoError(t, err)
	assert.Equal(t, "us-west-2", cfg.Region)
	assert.Equal(t, "arn:aws:sns:us-west-2:123:test-topic", cfg.TopicARN)
}

func TestNewConfigFromEnv_MissingRegion(t *testing.T) {
	t.Setenv("AWS_REGION", "")
	_, err := NewConfigFromEnv()
	assert.Error(t, err)
}
