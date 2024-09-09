package kafka

import (
	"context"
	"errors"
	"github.com/IBM/sarama"
	"gorm.io/gorm"
	"log"
	"mq-poc/model"
	"mq-poc/util"
	"os"
	"os/signal"
	"syscall"
)

func StartConsume(cfg *util.KafkaConfig, db *gorm.DB) {
	keepRunning := true
	config := sarama.NewConfig()
	config.Version = sarama.V3_6_0_0
	config.Consumer.Return.Errors = true

	dbq := model.NewDBQueue(db)
	dbq.Start()
	defer dbq.Stop()
	consumer := model.Consumer{
		Ready: make(chan bool),
		DbQ:   dbq,
	}

	// Create new consumer
	client, err := sarama.NewConsumerGroup(cfg.Brokers, "topics-customer-group", config)
	if err != nil {
		log.Fatalln("Failed to start Sarama consumer:", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Fatalln("Failed to close Sarama consumer:", err)
		}
	}()

	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	// Define the topic and partition to consume
	topics := []string{cfg.Topic}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := client.Consume(ctx, topics, &consumer); err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					return
				}
				log.Panicf("Error from consumer: %v", err)
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}
			consumer.Ready = make(chan bool)
		}
	}()

	<-consumer.Ready // Await till the consumer has been set up
	log.Println("Sarama consumer up and running!...")

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	for keepRunning {
		select {
		case <-ctx.Done():
			keepRunning = false
			log.Println("terminating: context cancelled")
		case <-sigterm:
			keepRunning = false
			log.Println("terminating: via signal")
		}
	}
	cancel()
	log.Println("terminating: bye!")
}

func StartConsumes(cfg *util.Kafka3Config, db *gorm.DB) {
	keepRunning := true
	config := sarama.NewConfig()
	config.Version = sarama.V3_6_0_0
	config.Consumer.Return.Errors = true

	dbq := model.NewDBQueue(db)
	dbq.Start()
	defer dbq.Stop()
	consumer := model.Consumer{
		Ready: make(chan bool),
		DbQ:   dbq,
	}

	// Create new consumer
	client, err := sarama.NewConsumerGroup(cfg.Brokers, "leg2-customer-group", config)
	if err != nil {
		log.Fatalln("Failed to start Sarama consumer:", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Fatalln("Failed to close Sarama consumer:", err)
		}
	}()

	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	// Define the topic and partition to consume
	topics := cfg.Topics
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := client.Consume(ctx, topics, &consumer); err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					return
				}
				log.Panicf("Error from consumer: %v", err)
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}
			consumer.Ready = make(chan bool)
		}
	}()

	<-consumer.Ready // Await till the consumer has been set up
	log.Println("Sarama consumer up and running!...")

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	for keepRunning {
		select {
		case <-ctx.Done():
			keepRunning = false
			log.Println("terminating: context cancelled")
		case <-sigterm:
			keepRunning = false
			log.Println("terminating: via signal")
		}
	}
	cancel()
	log.Println("terminating: bye!")
}
