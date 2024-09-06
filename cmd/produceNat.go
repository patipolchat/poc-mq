/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/spf13/cobra"
	"log"
	"mq-poc/model"
	"sync"
	"time"
)

// produceNatCmd represents the produceNat command
var produceNatCmd = &cobra.Command{
	Use:   "produceNat",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		nc, err := nats.Connect("nats://localhost:4222, nats://localhost:4223, nats://localhost:4224")
		if err != nil {
			log.Fatalln("Cannot Connect", err)
		}
		defer nc.Close()
		js, err := jetstream.New(nc)
		if err != nil {
			log.Fatalln("JetStreamErr", err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		_, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
			Name:     "foo",
			Subjects: []string{"foo"},
		})

		if err != nil {
			log.Fatalln("Cannot create stream", err)
		}

		randomStr, err := model.GenerateRandomString(1000)
		if err != nil {
			log.Fatalf("Failed to generate random string: %v", err)
		}

		for i := 0; i < 1; i++ {
			wg := &sync.WaitGroup{}
			for j := 0; j < 6; j++ {
				wg.Add(1)
				now := time.Now()
				msg := model.Message{
					ID:       uuid.NewString(),
					Start:    &now,
					End:      nil,
					Duration: nil,
					Tag:      "Nat1000-x",
					Data:     randomStr,
				}
				b, err := json.Marshal(msg)
				if err != nil {
					log.Fatal(err)
				}
				go func(b []byte) {
					defer wg.Done()
					_, err := js.PublishMsg(ctx, &nats.Msg{
						Subject: "foo",
						Data:    b,
					})
					if err != nil {
						log.Fatalln("cannot publish", err)
					}
				}(b)
			}
			wg.Wait()
			fmt.Println("Round", i+1)
			time.Sleep(5 * time.Second)
		}
		fmt.Println("Published message to NATS server")
	},
}

func init() {
	rootCmd.AddCommand(produceNatCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// produceNatCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// produceNatCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
