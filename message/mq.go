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

// 返回默认MQ单例
func Default() *MessageQueue {
	once.Do(func() { instance = New() })
	return instance
}

// 新建MQ
func New() *MessageQueue {
	return &MessageQueue{mq: make(map[string][]*Receiver), exit: make(chan bool)}
}

type Receiver struct {
	userID  int64              // 接收者ID
	channel chan string        // 消息载体
	cancel  context.CancelFunc //取消消息监听协程
	isGroup bool               // 是否为群聊
}

type MessageQueue struct {
	mq   map[string][]*Receiver //内部的订阅关系map
	exit chan bool              // 退出信号
	lock sync.RWMutex           // 保证mq协程安全的读写互斥锁
}

// 发信者只需要实现发送消息接口
type Messager interface {
	SendMessage(userID int64, msg string, isGroup bool)
}

// 监听消息，如接收到消息则通过messager发送给指定user
func (r *Receiver) checkMessage(bot Messager, ctx context.Context) {
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
func (m *MessageQueue) read(topic string) []*Receiver {
	m.lock.RLock()
	receiverList := m.mq[topic]
	m.lock.RUnlock()
	return receiverList
}

// 写消息队列
func (m *MessageQueue) write(topic string, receiverList ...*Receiver) {
	m.lock.Lock()
	m.mq[topic] = append(m.mq[topic], receiverList...)
	m.lock.Unlock()
}

// 覆写消息队列
func (m *MessageQueue) overwrite(topic string, receiverList []*Receiver) {
	m.lock.Lock()
	m.mq[topic] = receiverList
	m.lock.Unlock()
}

// 清空消息队列
func (m *MessageQueue) clear() {
	m.lock.Lock()
	m.mq = make(map[string][]*Receiver)
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

// 向消息队列推送消息
func (m *MessageQueue) Publish(topic, message string) error {
	if !m.checkStatus() {
		return errors.New("message queue closed")
	}
	receiverList := m.read(topic)
	m.broadcast(message, receiverList)
	return nil
}

// 内部的消息广播实现
func (m *MessageQueue) broadcast(message string, receiverList []*Receiver) {
	alert := func(receiver *Receiver) {
		select {
		case receiver.channel <- message:
		case <-time.After(10 * time.Second):
		case <-m.exit:
			return
		}
	}
	for _, receiver := range receiverList {
		go alert(receiver)
	}
}

// 订阅topic（在订阅同时启动订阅者消息监听，故需要bot实例）
func (m *MessageQueue) Subscribe(bot Messager, topic string, userID int64, isGroup bool) error {
	if !m.checkStatus() {
		return errors.New("message queue closed")
	}
	receiverList := m.read(topic)
	for _, receiver := range receiverList {
		if receiver.userID == userID {
			return errors.New("user already exist")
		}
	}
	channel := make(chan string, 1)
	ctx, cancel := context.WithCancel(context.Background())
	receiver := &Receiver{
		userID:  userID,
		channel: channel,
		cancel:  cancel,
		isGroup: isGroup,
	}
	m.write(topic, receiver)
	go receiver.checkMessage(bot, ctx)
	return nil
}

// 取消订阅topic
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

// 关闭消息队列
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

// 返回topic（即url）列表，供rss更新协程查询
func (m *MessageQueue) GetUrls() []string {
	var result []string
	m.lock.RLock()
	for url := range m.mq {
		result = append(result, url)
	}
	m.lock.RUnlock()
	return result
}
