package model

import (
	"fmt"
	"gorm.io/gorm"
	"log"
)

type DBQueue struct {
	qChan chan *Message
	db    *gorm.DB
	stop  chan bool
}

func NewDBQueue(db *gorm.DB) *DBQueue {
	return &DBQueue{
		qChan: make(chan *Message, 100),
		stop:  make(chan bool),
		db:    db,
	}
}

func (q *DBQueue) AddMsg(msg *Message) {
	q.qChan <- msg
}

func (q *DBQueue) Start() {
	go q.saveMsg()
}

func (q *DBQueue) Stop() {
	q.stop <- true
	log.Printf("Stop Queue\n")
}

func (q *DBQueue) saveMsg() {
	alive := true
	for alive {
		select {
		case msg := <-q.qChan:
			q.db.Create(msg)
		case ready := <-q.stop:
			if ready {
				close(q.qChan)
				alive = false
				break
			}
		}
	}
	fmt.Println("Stop Queue")
}
