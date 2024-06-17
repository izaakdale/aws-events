package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/izaakdale/aws-events/listener"
	"github.com/izaakdale/aws-events/publisher"
)

type EventType int

const (
	UNKNOWN_EVENT EventType = iota
	RESOURCE_CREATED
	RESOURCE_DELETED
)

type event struct {
	ID   int
	Type EventType
}

func main() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println(err)
		return
	}
	cfg.Credentials = aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
		return aws.Credentials{AccessKeyID: "test", SecretAccessKey: "test"}, nil
	})
	cfg.EndpointResolver = aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
		return aws.Endpoint{URL: "http://localhost:4566"}, nil
	})

	ls, err := listener.New(cfg, "http://sqs.us-east-1.localhost.localstack.cloud:4566/000000000000/test-queue")
	if err != nil {
		panic(err)
	}
	errCh := make(chan error)
	go ls.Listen(context.TODO(), process, errCh)

	pub, err := publisher.New(cfg, "arn:aws:sns:us-east-1:000000000000:test-notif")
	if err != nil {
		panic(err)
	}

	go func() {
		e := event{
			ID:   0,
			Type: RESOURCE_CREATED,
		}
		for {
			e.ID++
			b, err := json.Marshal(e)
			if err != nil {
				errCh <- err
			}
			_, err = pub.Publish(context.TODO(), b)
			if err != nil {
				errCh <- err
			}
			time.Sleep(5 * time.Second)
		}
	}()

	shCh := make(chan os.Signal, 1)
	signal.Notify(shCh, os.Interrupt)

	for {
		select {
		case err := <-errCh:
			log.Println(err)
		case <-shCh:
			return
		}
	}
}

func process(ctx context.Context, msg []byte) error {
	e := event{}
	err := json.Unmarshal(msg, &e)
	if err != nil {
		return err
	}

	switch e.Type {
	case RESOURCE_CREATED:
		fmt.Println("resource created")
	case RESOURCE_DELETED:
		fmt.Println("resource deleted")
	default:
		fmt.Println("unknown event")
	}

	return nil
}
