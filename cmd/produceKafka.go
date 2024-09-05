/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"log"
	"mq-poc/model"
	"strconv"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

// produceKafkaCmd represents the produceKafka command
var produceKafkaCmd = &cobra.Command{
	Use:   "produceKafka",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Define Kafka broker addresses. Make sure this matches your Kafka setup.

		// Configure the producer
		config := sarama.NewConfig()
		config.Producer.Return.Successes = true
		config.Producer.Retry.Max = 5
		config.Producer.RequiredAcks = sarama.WaitForAll
		config.Producer.Partitioner = sarama.NewRoundRobinPartitioner
		config.Producer.Compression = sarama.CompressionSnappy

		// Create a new sync producer
		producer, err := sarama.NewSyncProducer(brokers, config)
		if err != nil {
			log.Fatalf("Failed to start Sarama producer: %v", err)
		}
		defer func() {
			if err := producer.Close(); err != nil {
				log.Fatalln("Failed to close producer:", err)
			}
		}()

		// Define the topic to produce messages to
		topic := "topic0"
		randomStr, err := model.GenerateRandomString(128256)
		if err != nil {
			log.Fatalf("Failed to generate random string: %v", err)
		}
		for i := 0; i < 1; i++ {
			wg := sync.WaitGroup{}
			for j := 0; j < 1000; j++ {
				wg.Add(1)
				start := time.Now()
				msg := model.Message{
					ID:       uuid.NewString(),
					Start:    &start,
					End:      nil,
					Duration: nil,
					Tag:      "Kafka1000-3",
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
			time.Sleep(5 * time.Second)
		}

		log.Println("Produced 10000 messages")
	},
}

func init() {
	rootCmd.AddCommand(produceKafkaCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// produceKafkaCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// produceKafkaCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func TimeNow() []byte {
	return []byte(strconv.FormatInt(time.Now().UnixNano(), 10))
}

func DummyMsg(msg string, topic string) *sarama.ProducerMessage {
	return &sarama.ProducerMessage{
		Topic: topic,
		Headers: []sarama.RecordHeader{
			{Key: []byte("start"), Value: TimeNow()},
		},
		Value: sarama.StringEncoder(msg),
	}
}
