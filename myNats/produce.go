package myNats

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"log"
	"mq-poc/model"
	"mq-poc/util"
	"sync"
	"time"
)

func ProduceMsg(cfg *util.NatConfig, tag model.Tag) {
	nc, err := nats.Connect(cfg.Url)
	if err != nil {
		log.Fatalln("Cannot Connect", err)
	}
	defer nc.Close()
	js, err := jetstream.New(nc)
	if err != nil {
		log.Fatalln("JetStreamErr", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	_, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     "foo",
		Subjects: []string{cfg.Subject},
	})

	if err != nil {
		log.Fatalln("Cannot create stream", err)
	}

	randomStr, err := model.GenerateRandomString(tag.StrLen)
	if err != nil {
		log.Fatalf("Failed to generate random string: %v", err)
	}

	wg := &sync.WaitGroup{}
	for j := 0; j < tag.N; j++ {
		wg.Add(1)
		now := time.Now()
		msg := model.Message{
			ID:       uuid.NewString(),
			Start:    &now,
			End:      nil,
			Duration: nil,
			Tag:      tag.Name,
			Data:     randomStr,
		}
		b, err := json.Marshal(msg)
		if err != nil {
			log.Fatal(err)
		}
		go func(b []byte) {
			defer wg.Done()
			_, err := js.PublishMsg(ctx, &nats.Msg{
				Subject: cfg.Subject,
				Data:    b,
			})
			if err != nil {
				log.Fatalln("cannot publish", err)
			}
		}(b)
	}
	wg.Wait()

	fmt.Println("Published message to NATS server N: ", tag.N)
}
