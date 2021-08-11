package message

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	instance *MessageQueue
	once     sync.Once
)

func Default() *MessageQueue {
	once.Do(func() { instance = New() })
	return instance
}

func New() *MessageQueue {
	return &MessageQueue{mq: make(map[string][]Receiver), exit: make(chan bool)}
}

type Receiver struct {
	userID  int64
	channel chan string
	cancel  context.CancelFunc
	isGroup bool // 是否为群聊
}

type MessageQueue struct {
	mq   map[string][]Receiver
	exit chan bool
	lock sync.RWMutex
}

type messager interface {
	SendMessage(userID int64, msg string, isGroup bool)
}

func (r *Receiver) checkMessage(bot messager, ctx context.Context) {
	var msg string
	for {
		select {
		case msg = <-r.channel:
			go bot.SendMessage(r.userID, msg, r.isGroup)
		case <-ctx.Done():
			return
		}
	}
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

func (m *MessageQueue) Subscribe(bot messager, topic string, userID int64, isGroup bool) error {
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
		isGroup: isGroup,
	}
	m.write(topic, receiver)
	go receiver.checkMessage(bot, ctx)
	return nil
}

func (m *MessageQueue) Unsubscribe(topic string, userID int64) {
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
