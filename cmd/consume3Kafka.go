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

// consume3KafkaCmd represents the consume3Kafka command
var consume3KafkaCmd = &cobra.Command{
	Use:   "consume3Kafka",
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
		kafka.CreateManyTopic(cfg.Kafka3)
		kafka.StartConsumes(cfg.Kafka3, db)
	},
}

func init() {
	rootCmd.AddCommand(consume3KafkaCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// consume3KafkaCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// consume3KafkaCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
