package mq

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/amtoaer/bing-bong/client"
)

type Receiver struct {
	userID  string
	channel chan string
	cancel  context.CancelFunc
}

func (r *Receiver) checkMessage(ctx context.Context) {
	var msg string
	for {
		select {
		case msg = <-r.channel:
			go client.Robot.SendMessage(r.userID, msg)
		case <-ctx.Done():
			return
		}
	}
}

type MessageQueue struct {
	mq   map[string][]Receiver
	exit chan bool
	lock sync.RWMutex
}

// 读消息队列
func (m *MessageQueue) read(topic string) []Receiver {
	m.lock.RLock()
	receiverList := m.mq[topic]
	m.lock.RUnlock()
	return receiverList
}

// 写消息队列
func (m *MessageQueue) write(topic string, receiverList ...Receiver) {
	m.lock.Lock()
	m.mq[topic] = append(m.mq[topic], receiverList...)
	m.lock.Unlock()
}

// 覆写消息队列
func (m *MessageQueue) overwrite(topic string, receiverList []Receiver) {
	m.lock.Lock()
	m.mq[topic] = receiverList
	m.lock.Unlock()
}

// 清空消息队列
func (m *MessageQueue) clear() {
	m.lock.Lock()
	m.mq = make(map[string][]Receiver)
	m.lock.Unlock()
}

// 判断消息队列是否关闭
func (m *MessageQueue) checkStatus() bool {
	select {
	case <-m.exit:
		return false
	default:
		return true
	}
}

func (m *MessageQueue) Publish(topic, message string) error {
	if !m.checkStatus() {
		return errors.New("message queue closed")
	}
	receiverList := m.read(topic)
	m.broadcast(message, receiverList)
	return nil
}

func (m *MessageQueue) broadcast(message string, receiverList []Receiver) {
	alert := func(receiver Receiver) {
		select {
		case <-receiver.channel:
		case <-time.After(10 * time.Second):
		case <-m.exit:
			return
		}
	}
	for _, receiver := range receiverList {
		go alert(receiver)
	}
}

func (m *MessageQueue) Subscribe(topic string, userID string) error {
	receiverList := m.read(topic)
	for _, receiver := range receiverList {
		if receiver.userID == userID {
			return errors.New("user already exist")
		}
	}
	channel := make(chan string, 1)
	ctx, cancel := context.WithCancel(context.Background())
	receiver := Receiver{
		userID:  userID,
		channel: channel,
		cancel:  cancel,
	}
	m.write(topic, receiver)
	go receiver.checkMessage(ctx)
	return nil
}

func (m *MessageQueue) Unsubscribe(topic string, userID string) {
	if !m.checkStatus() {
		return
	}
	receiverList := m.read(topic)
	for index, receiver := range receiverList {
		if receiver.userID == userID {
			receiver.cancel()
			m.overwrite(topic, append(receiverList[:index], receiverList[index+1:]...))
			return
		}
	}
}

func (m *MessageQueue) Close() {
	if !m.checkStatus() {
		return
	}
	close(m.exit)
	for _, receiverList := range m.mq {
		for _, receiver := range receiverList {
			receiver.cancel()
		}
	}
	m.clear()
}
