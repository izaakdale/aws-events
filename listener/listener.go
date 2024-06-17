package listener

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

var (
	// match test events from AWS policy checher
	testEventMatcher        = regexp.MustCompile(`.*:TestEvent`)
	ErrClientNotInitialised = errors.New("uninitialised client")
)

const (
	defaultMaxNumberOfMessages = int32(10)
	defaultVisibiltyTimeout    = int32(5)
	defaultWaitTimeSeconds     = int32(10)
)

type (
	Client struct {
		sqsClient      sqsConsumeAPI
		input          *sqs.ReceiveMessageInput
		wantTestEvents bool
	}
	sqsConsumeAPI interface {
		ReceiveMessage(ctx context.Context,
			params *sqs.ReceiveMessageInput,
			optFns ...func(*sqs.Options),
		) (*sqs.ReceiveMessageOutput, error)
		DeleteMessage(ctx context.Context,
			params *sqs.DeleteMessageInput,
			optFns ...func(*sqs.Options),
		) (*sqs.DeleteMessageOutput, error)
	}
	configOptions struct {
		maxNumberOfMessages *int32
		visibilityTimeout   *int32
		waitTimeSeconds     *int32
		attributeNames      []types.MessageSystemAttributeName
		wantTestEvents      *bool
	}
	option func(opt *configOptions) error
	// This function allows the user to define how their messages should be processed.
	ProcessorFunc func(context.Context, []byte) error
)

// New creates a new SQS listener listening to the queue with the specifed options.
func New(cfg aws.Config, queueURL string, optFuncs ...option) (*Client, error) {
	var options configOptions
	for _, optFunc := range optFuncs {
		err := optFunc(&options)
		if err != nil {
			return nil, err
		}
	}

	var cli Client
	cli.input = &sqs.ReceiveMessageInput{
		QueueUrl: &queueURL,
	}

	if options.attributeNames != nil {
		cli.input.MessageSystemAttributeNames = options.attributeNames
	} else {
		cli.input.MessageSystemAttributeNames = []types.MessageSystemAttributeName{
			types.MessageSystemAttributeNameAll,
		}
	}

	if options.maxNumberOfMessages != nil {
		cli.input.MaxNumberOfMessages = *options.maxNumberOfMessages
	} else {
		cli.input.MaxNumberOfMessages = defaultMaxNumberOfMessages
	}
	if options.visibilityTimeout != nil {
		cli.input.VisibilityTimeout = *options.visibilityTimeout
	} else {
		cli.input.VisibilityTimeout = defaultVisibiltyTimeout
	}
	if options.waitTimeSeconds != nil {
		cli.input.WaitTimeSeconds = *options.waitTimeSeconds
	} else {
		cli.input.WaitTimeSeconds = defaultWaitTimeSeconds
	}

	if options.wantTestEvents != nil {
		cli.wantTestEvents = *options.wantTestEvents
	} else {
		cli.wantTestEvents = false
	}

	cli.sqsClient = sqs.NewFromConfig(cfg)
	return &cli, nil
}

// Listen triggers a never ending for loop that continually requests the specified queue for messages.
func (client *Client) Listen(ctx context.Context, pf ProcessorFunc, errChan chan<- error) {
	if client == nil {
		errChan <- ErrClientNotInitialised
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msgResult, err := client.sqsClient.ReceiveMessage(ctx, client.input)
			if err != nil {
				errChan <- fmt.Errorf("failed to receive message: %w", err)
			}

			if msgResult == nil {
				continue
			}

			for _, m := range msgResult.Messages {
				awsEvent := struct {
					Type    string `json:"Event"`
					Message string `json:"Message"`
				}{}
				if err := json.Unmarshal([]byte(*m.Body), &awsEvent); err != nil {
					errChan <- fmt.Errorf("failed to unmarshal message: %w", err)
					continue
				}

				if client.wantTestEvents {
					err = pf(ctx, []byte(awsEvent.Message))
					if err != nil {
						errChan <- err
					}
				} else {

					if !testEventMatcher.MatchString(awsEvent.Type) {
						err = pf(ctx, []byte(awsEvent.Message))
						if err != nil {
							errChan <- err
						}
					}
				}

				dMInput := &sqs.DeleteMessageInput{
					QueueUrl:      client.input.QueueUrl,
					ReceiptHandle: m.ReceiptHandle,
				}
				_, err = client.sqsClient.DeleteMessage(ctx, dMInput)
				if err != nil {
					errChan <- err
				}
			}
		}
	}
}
