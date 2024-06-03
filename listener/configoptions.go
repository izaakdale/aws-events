package listener

import "github.com/aws/aws-sdk-go-v2/service/sqs/types"

// WithMaxNumerOfMessages dictaces how many messages can be returned from a queue in one go.
// Defaults to 10.
func WithMaxNumerOfMessages(n int32) option {
	return func(opt *configOptions) error {
		opt.maxNumberOfMessages = &n
		return nil
	}
}

// WithVisibilityTimeout dictates how long a queue message will be hidden from other clients.
// Defaults to 5 seconds.
func WithVisibilityTimeout(v int32) option {
	return func(opt *configOptions) error {
		opt.visibilityTimeout = &v
		return nil
	}
}

// WithWaitTimeSeconds dictates how long the request to the queue will wait for a message.
// Defaults to 10 seconds
func WithWaitTimeSeconds(s int32) option {
	return func(opt *configOptions) error {
		opt.waitTimeSeconds = &s
		return nil
	}
}

// WithAttributeNames dictates which attributes to return from the queue.
func WithAttributeNames(n []types.MessageSystemAttributeName) option {
	return func(opt *configOptions) error {
		opt.attributeNames = n
		return nil
	}
}

func WithTestEvents(want bool) option {
	return func(opt *configOptions) error {
		opt.wantTestEvents = &want
		return nil
	}
}
