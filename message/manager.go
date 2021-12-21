package message

import (
	"sync"

	"github.com/amtoaer/bing-bong/model"
)

var (
	once sync.Once
	mm   *Manager
)

type Manager struct {
	mq  *Queue
	rm  *ReceiverManager
	bot Messager
}

func NewManager() *Manager {
	return &Manager{
		mq: NewMessageQueue(),
		rm: NewReceiverManager(),
	}
}

func DefaultManager() *Manager {
	once.Do(func() {
		mm = NewManager()
	})
	return mm
}

func (m *Manager) Publish(topic, message string) error {
	return m.mq.Publish(topic, message)
}

func (m *Manager) Subscribe(topic string, user *model.User) error {
	var (
		receiver *Receiver
		ok       bool
	)
	if receiver, ok = m.rm.GetReceiver(user.Account, user.IsGroup); !ok {
		receiver = m.rm.RegisterReceiver(user.Account, user.IsGroup)
		go receiver.CheckMessage(m.bot)
	}
	return m.mq.Subscribe(topic, receiver.channel)
}

func (m *Manager) UnSubscribe(topic string, user *model.User) error {
	if receiver, ok := m.rm.GetReceiver(user.Account, user.IsGroup); ok {
		m.mq.Unsubscribe(topic, receiver.channel)
	}
	return nil
}

func (m *Manager) GetTopics() []string {
	return m.mq.GetTopics()
}

func (m *Manager) RegisterBot(bot Messager) {
	m.bot = bot
}
