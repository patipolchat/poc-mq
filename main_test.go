package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"mq-poc/kafka"
	"mq-poc/model"
	"mq-poc/myNats"
	"mq-poc/util"
	"testing"
	"time"
)

type ProduceSuite struct {
	suite.Suite
	cfg *util.Config
	db  *gorm.DB
}

func (suite *ProduceSuite) SetupSuite() {
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()
	suite.Require().NoError(err)

	cfg, err := util.NewConfig()
	suite.Require().NoError(err)

	suite.cfg = cfg
	suite.db, err = util.GetDB(cfg.DB)
	suite.Require().NoError(err)

	kafka.CreateTopic(cfg.Kafka)
}

func (suite *ProduceSuite) TestProduceKafka() {
	tag := model.Tag{
		Name:   "Kafka-" + uuid.NewString(),
		StrLen: 2048,
		N:      10000,
	}
	kafka.ProduceMsg(suite.cfg.Kafka, suite.db, tag)
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

func (suite *ProduceSuite) TestProduceNats() {
	tag := model.Tag{
		Name:   "Nats-" + uuid.NewString(),
		StrLen: 2048,
		N:      10000,
	}
	myNats.ProduceMsg(suite.cfg.Nats, tag)
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

func (suite *ProduceSuite) countMessages(tag model.Tag) (int64, error) {
	var count int64
	err := suite.db.Model(&model.Message{}).Where(`tag=?`, tag.Name).Count(&count).Error
	if err != nil {
		return 0, err
	}
	fmt.Println("Count", count)

	return count, nil
}

func (suite *ProduceSuite) getAvg(tag model.Tag) (string, error) {
	var avg string
	err := suite.db.Model(&model.Message{}).Where(`tag=?`, tag.Name).Select("AVG(duration)").Row().Scan(&avg)
	if err != nil {
		return "", err
	}

	return avg, nil
}

func TestProduceSuite(t *testing.T) {
	suite.Run(t, new(ProduceSuite))
}
