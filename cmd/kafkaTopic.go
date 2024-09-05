/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/IBM/sarama"
	"log"

	"github.com/spf13/cobra"
)

// kafkaTopicCmd represents the kafkaTopic command
var kafkaTopicCmd = &cobra.Command{
	Use:   "kafkaTopic",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		CreateTopic()
		// Configure Sarama
		config := sarama.NewConfig()
		config.Version = sarama.V3_6_0_0 // Use the Kafka version you are running

		// Create a new ClusterAdmin client
		admin, err := sarama.NewClusterAdmin(broker, config)
		if err != nil {
			log.Fatalf("Error creating cluster admin: %v", err)
		}
		defer func() {
			if err := admin.Close(); err != nil {
				log.Fatalf("Error closing cluster admin: %v", err)
			}
		}()

		// Define the topic details
		topicName := "topic0"
		numPartitions := 3
		replicationFactor := int16(1)
		topicDetail := &sarama.TopicDetail{
			NumPartitions:     int32(numPartitions),
			ReplicationFactor: replicationFactor,
			ConfigEntries:     map[string]*string{},
		}

		// Create the new topic
		err = admin.CreateTopic(topicName, topicDetail, false)
		if err != nil {
			log.Fatalf("Error creating topic: %v", err)
		} else {
			fmt.Printf("Topic '%s' created successfully\n", topicName)
		}
	},
}

func init() {
	rootCmd.AddCommand(kafkaTopicCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// kafkaTopicCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// kafkaTopicCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
