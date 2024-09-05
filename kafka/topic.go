package kafka

import (
	"fmt"
	"github.com/IBM/sarama"
	"log"
	"mq-poc/util"
)

func CreateTopic(cfg *util.KafkaConfig) error {
	// Configure Sarama
	config := sarama.NewConfig()
	config.Version = sarama.V3_6_0_0 // Use the Kafka version you are running

	// Create a new ClusterAdmin client
	admin, err := sarama.NewClusterAdmin(cfg.Brokers, config)
	if err != nil {
		log.Fatalf("Error creating cluster admin: %v", err)
	}
	defer func() {
		if err := admin.Close(); err != nil {
			log.Fatalf("Error closing cluster admin: %v", err)
		}
	}()

	// Define the topic details
	replicationFactor := int16(1)
	topicDetail := &sarama.TopicDetail{
		NumPartitions:     int32(cfg.Partition),
		ReplicationFactor: replicationFactor,
		ConfigEntries:     map[string]*string{},
	}

	// Create the new topic
	err = admin.CreateTopic(cfg.Topic, topicDetail, false)
	if err != nil {
		log.Printf("Error creating topic: %v\n", err)
		return err
	} else {
		fmt.Printf("Topic '%s' created successfully\n", cfg.Topic)
		return nil
	}
}
