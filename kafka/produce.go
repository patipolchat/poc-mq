package kafka

import (
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"log"
	"mq-poc/model"
	"mq-poc/util"
	"sync"
	"time"
)

func ProduceMsg(cfg *util.KafkaConfig, db *gorm.DB, tag model.Tag) {
	// Configure the producer
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Retry.Max = 5
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	config.Producer.Compression = sarama.CompressionSnappy

	// Create a new sync producer
	producer, err := sarama.NewSyncProducer(cfg.Brokers, config)
	if err != nil {
		log.Fatalf("Failed to start Sarama producer: %v", err)
	}
	defer func() {
		if err := producer.Close(); err != nil {
			log.Fatalln("Failed to close producer:", err)
		}
	}()

	// Define the topic to produce messages to
	topic := cfg.Topic
	randomStr, err := model.GenerateRandomString(tag.StrLen)
	if err != nil {
		log.Fatalf("Failed to generate random string: %v", err)
	}
	for i := 0; i < 1; i++ {
		wg := sync.WaitGroup{}
		for j := 0; j < tag.N; j++ {
			wg.Add(1)
			start := time.Now()
			msg := model.Message{
				ID:       uuid.NewString(),
				Start:    &start,
				End:      nil,
				Duration: nil,
				Tag:      tag.Name,
				Data:     randomStr,
			}
			b, err := json.Marshal(msg)
			if err != nil {
				log.Fatalf("Failed to marshal message: %v", err)
			}
			go func(b []byte) {
				defer wg.Done()
				_, _, err = producer.SendMessage(&sarama.ProducerMessage{
					Topic: topic,
					Value: sarama.ByteEncoder(b),
				})
				if err != nil {
					log.Printf("Failed to send message: %v", err)
				}
			}(b)
		}
		log.Println("Producing 10000 messages")
		wg.Wait()
	}

	log.Println("Produced 10000 messages")
}
