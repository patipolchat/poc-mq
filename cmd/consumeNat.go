/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/spf13/cobra"
	"log"
	"mq-poc/model"
	"time"
)

// consumeNatCmd represents the consumeNat command
var consumeNatCmd = &cobra.Command{
	Use:   "consumeNat",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		db := GetDB()

		q := model.NewDBQueue(db)
		q.Start()
		nc, err := nats.Connect("nats://nats:4222")
		if err != nil {
			log.Fatal(err)
		}
		defer nc.Close()
		js, err := jetstream.New(nc)
		if err != nil {
			log.Fatalln("JetStreamErr", err)
		}
		ctx := context.Background()
		stream, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
			Name:     "foo",
			Subjects: []string{"foo"},
		})
		if err != nil {
			log.Fatalf("Failed to create stream: %v", err)
		}

		consumer, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
			Name:    "foo",
			Durable: "foo",
		})
		if err != nil {
			log.Fatalf("Failed to create consumer: %v", err)
		}
		fmt.Println("Connected to NATS server")
		cc, err := consumer.Consume(func(msg jetstream.Msg) {
			defer msg.Ack()
			end := time.Now()
			go func(msg jetstream.Msg, end time.Time) {
				var message model.Message
				err := json.Unmarshal(msg.Data(), &message)
				if err != nil {
					log.Fatalf("Failed to parse message: %v", err)
				}
				message.End = &end
				message.SetDuration()
				q.AddMsg(&message)
			}(msg, end)
		})
		defer cc.Stop()
		for {
			select {
			case <-cc.Closed():
				log.Println("Consumer closed")
				return
			}
		}
		//ch := make(chan *nats.Msg, 64)
		//sub, err := nc.ChanSubscribe("foo", ch)
		//if err != nil {
		//	log.Fatalf("Failed to subscribe to NATS server: %v", err)
		//}
		//defer sub.Unsubscribe()
		//ctx, cancel := context.WithCancel(context.Background())
		//defer cancel()
		//for {
		//	select {
		//	case msg := <-ch:
		//		end := time.Now()
		//		go func(msg *nats.Msg, end time.Time) {
		//			var message model.Message
		//			err := json.Unmarshal(msg.Data, &message)
		//			if err != nil {
		//				log.Fatalf("Failed to parse message: %v", err)
		//			}
		//			message.End = &end
		//			message.SetDuration()
		//			q.AddMsg(&message)
		//		}(msg, end)
		//	case <-ctx.Done():
		//		close(ch)
		//		return
		//	}
		//}

	},
}

func init() {
	rootCmd.AddCommand(consumeNatCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// consumeNatCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// consumeNatCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
