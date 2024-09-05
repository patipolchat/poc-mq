/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/IBM/sarama"
	"github.com/spf13/cobra"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"mq-poc/model"
	"os"
)

var broker = []string{"kafka-0:29092", "kafka-1:29092", "kafka-2:29092"}

var brokers = []string{"localhost:9092", "localhost:9093", "localhost:9094"}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mq-poc",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mq-poc.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

func GetDB() *gorm.DB {
	dsn := "host=postgres user=citizix_user password=S3cret dbname=citizix_db port=5432 sslmode=disable TimeZone=Asia/Bangkok"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&model.Message{})
	return db
}

func CreateTopic() {
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
}
