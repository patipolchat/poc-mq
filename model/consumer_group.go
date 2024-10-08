package model

import (
	"encoding/json"
	"github.com/IBM/sarama"
	"log"
	"time"
)

type Consumer struct {
	Ready     chan bool
	StartTime time.Time
	DbQ       *DBQueue
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	consumer.StartTime = time.Now()
	close(consumer.Ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
// Once the Messages() channel is closed, the Handler must finish its processing
// loop and exit.
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/IBM/sarama/blob/main/consumer_group.go#L27-L29
	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				log.Printf("message channel was closed")
				return nil
			}
			endTime := time.Now()
			go func(msg *sarama.ConsumerMessage, endTime time.Time) {
				var message Message
				err := json.Unmarshal(msg.Value, &message)
				if err != nil {
					log.Fatalf("Failed to parse message: %v", err)
				}
				message.End = &endTime
				message.SetDuration()
				message.Data = ""
				consumer.DbQ.AddMsg(&message)
			}(msg, endTime)
			session.MarkMessage(msg, "")
		// Should return when `session.Context()` is done.
		// If not, will raise `ErrRebalanceInProgress` or `read tcp <ip>:<port>: i/o timeout` when kafka rebalance. see:
		// https://github.com/IBM/sarama/issues/1192
		case <-session.Context().Done():
			log.Printf("session context was done")
			return nil
		}
	}
}
