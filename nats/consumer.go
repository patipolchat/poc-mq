package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"gorm.io/gorm"
	"log"
	"mq-poc/model"
	"mq-poc/util"
	"time"
)

func StartConsume(cfg *util.NatConfig, db *gorm.DB) {
	q := model.NewDBQueue(db)
	q.Start()
	nc, err := nats.Connect("nats://nats:4222")
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()
	js, err := jetstream.New(nc)
	if err != nil {
		log.Fatalln("JetStreamErr", err)
	}
	ctx := context.Background()
	stream, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     "foo",
		Subjects: []string{"foo"},
	})
	if err != nil {
		log.Fatalf("Failed to create stream: %v", err)
	}

	consumer, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Name:    "foo",
		Durable: "foo",
	})
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}
	fmt.Println("Connected to NATS server")
	cc, err := consumer.Consume(func(msg jetstream.Msg) {
		defer msg.Ack()
		end := time.Now()
		go func(msg jetstream.Msg, end time.Time) {
			var message model.Message
			err := json.Unmarshal(msg.Data(), &message)
			if err != nil {
				log.Fatalf("Failed to parse message: %v", err)
			}
			message.End = &end
			message.SetDuration()
			q.AddMsg(&message)
		}(msg, end)
	})
	defer cc.Stop()
	for {
		select {
		case <-cc.Closed():
			log.Println("Consumer closed")
			return
		}
	}
}
