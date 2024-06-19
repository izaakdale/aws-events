package publisher

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

//go:generate mockgen -destination mocks/mock_publisher.go -package mocks . SNSPublishAPI

var (
	ErrClientNotInitialised = errors.New("uninitialised client")
)

type (
	Client struct {
		sns      SNSPublishAPI
		TopicArn string
	}
	SNSPublishAPI interface {
		Publish(
			ctx context.Context,
			params *sns.PublishInput,
			optFns ...func(*sns.Options),
		) (*sns.PublishOutput, error)
	}
	configOptions struct {
		publisher SNSPublishAPI
	}
	option func(opt *configOptions) error
)

func New(cfg aws.Config, topicArn string, optFuncs ...option) (*Client, error) {
	var options configOptions
	for _, optFunc := range optFuncs {
		err := optFunc(&options)
		if err != nil {
			return nil, err
		}
	}
	if options.publisher != nil {
		client := &Client{
			TopicArn: topicArn,
			sns:      options.publisher,
		}
		return client, nil
	}
	return &Client{
		TopicArn: topicArn,
		sns:      sns.NewFromConfig(cfg),
	}, nil
}

// Publish sends a message to the Topic initialised in the client.
// Returns the message id and an error
func (client *Client) Publish(ctx context.Context, msg []byte) (*string, error) {
	input := &sns.PublishInput{
		Message:  aws.String(string(msg)),
		TopicArn: aws.String(client.TopicArn),
	}
	result, err := client.sns.Publish(ctx, input)
	if err != nil {
		return nil, err
	}
	return result.MessageId, nil
}

// WithPublisher allows the client to use their own publisher with the package.
func WithPublisher(p SNSPublishAPI) option {
	return func(opt *configOptions) error {
		opt.publisher = p
		return nil
	}
}
