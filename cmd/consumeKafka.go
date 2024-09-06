/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"mq-poc/kafka"
	"mq-poc/util"
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
		viper.AddConfigPath(".")
		viper.SetConfigName(cfgName)
		viper.SetConfigType("yaml")
		err := viper.ReadInConfig()
		if err != nil {
			log.Fatalln("Cannot read config", err)
		}
		cfg, err := util.NewConfig()
		if err != nil {
			log.Fatalln("Cannot Get config", err)
		}
		db, err := util.GetDB(cfg.DB)
		if err != nil {
			log.Fatalln("Cannot Get DB", err)
		}
		kafka.CreateTopic(cfg.Kafka)
		kafka.StartConsume(cfg.Kafka, db)
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
