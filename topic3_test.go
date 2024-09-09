package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"log"
	"mq-poc/kafka"
	"mq-poc/model"
	"mq-poc/util"
	"sync"
	"testing"
	"time"
)

const strLen = 2048

type Topic3Suite struct {
	suite.Suite
	cfg      *util.Config
	producer sarama.SyncProducer
	db       *gorm.DB
}

func (suite *Topic3Suite) SetupSuite() {
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()
	suite.Require().NoError(err)

	cfg, err := util.NewConfig()
	suite.Require().NoError(err)

	suite.db, err = util.GetDB(cfg.DB)
	suite.Require().NoError(err)

	suite.cfg = cfg
	suite.Require().NoError(err)

	kafka.CreateManyTopic(cfg.Kafka3)

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Retry.Max = 5
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	config.Producer.Compression = sarama.CompressionSnappy

	// Create a new sync producer
	producer, err := sarama.NewSyncProducer(suite.cfg.Kafka3.Brokers, config)
	suite.Require().NoError(err)
	suite.producer = producer
}

func (suite *Topic3Suite) TearDownSuite() {
	suite.Require().NoError(suite.producer.Close())
}

func (suite *Topic3Suite) TestProduceKafka() {
	tag := model.Tag{
		Name:   "Kafka3-" + uuid.NewString(),
		StrLen: strLen,
		N:      10000,
	}
	suite.ProduceKafka(tag)

	var count int64
	ch := time.Tick(time.Second * 3)
	for range ch {
		var err error
		count, err = suite.countMessages(tag)
		suite.Require().NoError(err)
		if count >= int64(tag.N) {
			break
		}
	}
	avg, err := suite.getAvg(tag)
	suite.Require().NoError(err)
	fmt.Printf("Tag %s Avg %d: %s\n", tag.Name, tag.N, avg)
}

func (suite *Topic3Suite) TestProduceNats() {
	tag := model.Tag{
		Name:   "Nats3-" + uuid.NewString(),
		StrLen: strLen,
		N:      100,
	}
	suite.ProduceNats(tag)

	var count int64
	ch := time.Tick(time.Second * 3)
	for range ch {
		var err error
		count, err = suite.countMessages(tag)
		suite.Require().NoError(err)
		if count >= int64(tag.N) {
			break
		}
	}
	avg, err := suite.getAvg(tag)
	suite.Require().NoError(err)
	fmt.Printf("Tag %s Avg %d: %s\n", tag.Name, tag.N, avg)
}

func (suite *Topic3Suite) ProduceNats(tag model.Tag) {
	cfg := suite.cfg.Nat3
	nc, err := nats.Connect(cfg.Url)
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

	randomStr, err := model.GenerateRandomString(tag.StrLen)
	if err != nil {
		log.Fatalf("Failed to generate random string: %v", err)
	}

	wg := &sync.WaitGroup{}
	subjectI := 0
	for j := 0; j < tag.N; j++ {
		wg.Add(1)
		now := time.Now()
		msg := model.Message{
			ID:       uuid.NewString(),
			Start:    &now,
			End:      nil,
			Duration: nil,
			Tag:      tag.Name,
			Data:     randomStr,
		}
		b, err := json.Marshal(msg)
		if err != nil {
			log.Fatal(err)
		}

		subject := cfg.Subjects[subjectI]
		go func(subject string, b []byte, wg *sync.WaitGroup) {
			defer wg.Done()
			_, err := js.PublishMsg(ctx, &nats.Msg{
				Subject: subject,
				Data:    b,
			})
			if err != nil {
				log.Fatalln("cannot publish", err)
			}
		}(subject, b, wg)

		subjectI++
		if subjectI >= len(cfg.Subjects) {
			subjectI = 0
		}
	}
	log.Println("Wait all messages")
	wg.Wait()

	fmt.Println("Published message to NATS server N: ", tag.N)
}

func (suite *Topic3Suite) ProduceKafka(tag model.Tag) {
	// Define the topic to produce messages to
	randomStr, err := model.GenerateRandomString(tag.StrLen)
	if err != nil {
		log.Fatalf("Failed to generate random string: %v", err)
	}
	wg := sync.WaitGroup{}
	topicI := 0
	for i := 0; i < tag.N; i++ {
		wg.Add(1)
		start := time.Now()
		msg := model.Message{
			ID:       uuid.NewString(),
			Start:    &start,
			End:      nil,
			Duration: nil,
			Tag:      tag.Name,
			Data:     randomStr,
		}
		b, err := json.Marshal(msg)
		suite.Require().NoError(err)

		topic := suite.cfg.Kafka3.Topics[topicI]

		go func(topic string, b []byte) {
			defer wg.Done()
			_, _, err = suite.producer.SendMessage(&sarama.ProducerMessage{
				Topic: topic,
				Value: sarama.ByteEncoder(b),
			})
			if err != nil {
				log.Printf("Failed to send message: %v", err)
			}
		}(topic, b)
		topicI++
		if topicI >= len(suite.cfg.Kafka3.Topics) {
			topicI = 0
		}
	}
	wg.Wait()
	log.Printf("Produced %d messages success\n", tag.N)
}

func (suite *Topic3Suite) countMessages(tag model.Tag) (int64, error) {
	var count int64
	err := suite.db.Model(&model.Message{}).Where(`tag=?`, tag.Name).Count(&count).Error
	if err != nil {
		return 0, err
	}
	fmt.Println("Count", count)

	return count, nil
}

func (suite *Topic3Suite) getAvg(tag model.Tag) (string, error) {
	var avg string
	err := suite.db.Model(&model.Message{}).Where(`tag=?`, tag.Name).Select("AVG(duration)").Row().Scan(&avg)
	if err != nil {
		return "", err
	}

	return avg, nil
}

func TestTopic3Suite(t *testing.T) {
	suite.Run(t, new(Topic3Suite))
}
