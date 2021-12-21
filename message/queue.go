package message

import (
	"sync"
	"time"
)

type Queue struct {
	mq   map[string][]chan string //内部的订阅关系map
	lock sync.RWMutex             // 保证mq协程安全的读写互斥锁
}

func NewMessageQueue() *Queue {
	return &Queue{
		mq: make(map[string][]chan string),
	}
}

// 读对应topic的订阅者列表
func (m *Queue) read(topic string) []chan string {
	m.lock.RLock()
	receiverList := m.mq[topic]
	m.lock.RUnlock()
	return receiverList
}

// 向对应topic追加订阅者
func (m *Queue) write(topic string, receiverList ...chan string) {
	m.lock.Lock()
	m.mq[topic] = append(m.mq[topic], receiverList...)
	m.lock.Unlock()
}

// 覆写对应topic的订阅者
func (m *Queue) overwrite(topic string, receiverList []chan string) {
	m.lock.Lock()
	m.mq[topic] = receiverList
	m.lock.Unlock()
}

// 内部的消息广播实现
func (m *Queue) broadcast(message string, receiverList []chan string) {
	alert := func(receiver chan string) {
		select {
		case receiver <- message:
		case <-time.After(10 * time.Second):
			return
		}
	}
	for _, receiver := range receiverList {
		go alert(receiver)
	}
}

// 向topic推送消息
func (m *Queue) Publish(topic, message string) error {
	receiverList := m.read(topic)
	m.broadcast(message, receiverList)
	return nil
}

// 订阅topic
func (m *Queue) Subscribe(topic string, channel chan string) error {
	m.write(topic, channel)
	return nil
}

// 取消订阅topic
func (m *Queue) Unsubscribe(topic string, channel chan string) error {
	receiverList := m.read(topic)
	for index := range receiverList {
		receiverChan := receiverList[index]
		if receiverChan == channel {
			m.overwrite(topic, append(receiverList[:index], receiverList[index+1:]...))
		}
	}
	return nil
}

// 返回topic（即url）列表，供rss更新协程查询
func (m *Queue) GetTopics() []string {
	var result []string
	m.lock.RLock()
	for url := range m.mq {
		result = append(result, url)
	}
	m.lock.RUnlock()
	return result
}
