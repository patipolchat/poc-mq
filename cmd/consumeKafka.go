/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"errors"
	"github.com/IBM/sarama"
	"github.com/spf13/cobra"
	"log"
	"mq-poc/model"
	"os"
	"os/signal"
	"syscall"
)

// consumeKafkaCmd represents the consumeKafka command
var consumeKafkaCmd = &cobra.Command{
	Use:   "consumeKafka",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		keepRunning := true
		config := sarama.NewConfig()
		config.Version = sarama.V3_6_0_0
		config.Consumer.Return.Errors = true

		db := GetDB()

		dbq := model.NewDBQueue(db)
		dbq.Start()
		defer dbq.Stop()
		consumer := model.Consumer{
			Ready: make(chan bool),
			DbQ:   dbq,
		}

		// Create new consumer
		client, err := sarama.NewConsumerGroup(broker, "leg2-customer-group", config)
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
		topics := []string{"topic0"}
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
	},
}

func init() {
	rootCmd.AddCommand(consumeKafkaCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// consumeKafkaCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// consumeKafkaCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
