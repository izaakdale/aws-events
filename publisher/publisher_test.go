package publisher_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/golang/mock/gomock"

	"github.com/izaakdale/aws-events/publisher"
	"github.com/izaakdale/aws-events/publisher/mocks"
)

const (
	testTopicArn = "arn:aws:sns:us-east-1:000000000000:test-notif"
)

var (
	testID  = "testID"
	errTest = errors.New("test error")
)

func TestSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	pub := mocks.NewMockSNSPublishAPI(ctrl)

	pub.EXPECT().Publish(gomock.Any(), gomock.Any(), gomock.Len(0)).Return(&sns.PublishOutput{MessageId: &testID}, nil)
	client, err := publisher.New(aws.Config{}, testTopicArn, publisher.WithPublisher(pub))
	if err != nil {
		t.Fatal(err)
	}
	id, err := client.Publish(context.TODO(), []byte("test"))
	if err != nil {
		t.Fatal(err)
	}
	if id == nil {
		t.Fatal("id is nil")
	}
	if *id != testID {
		t.Errorf("expected %s, got %s", testID, *id)
	}
}

func TestFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	pub := mocks.NewMockSNSPublishAPI(ctrl)

	pub.EXPECT().Publish(gomock.Any(), gomock.Any(), gomock.Len(0)).Return(nil, errTest)
	client, err := publisher.New(aws.Config{}, testTopicArn, publisher.WithPublisher(pub))
	if err != nil {
		t.Fatal(err)
	}
	id, err := client.Publish(context.TODO(), []byte("test"))
	if err == nil {
		t.Fatal(err)
	}
	if id != nil {
		t.Fatal("id is nil")
	}
	if !errors.Is(err, errTest) {
		t.Errorf("expected error %s", errTest)
	}
}
