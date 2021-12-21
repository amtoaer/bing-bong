package message

import (
	"fmt"
	"sync"
)

// 发信者只需要实现发送消息接口
type Messager interface {
	SendMessage(userID int64, msg string, isGroup bool)
}

type Receiver struct {
	userID  int64       // 接收者ID
	isGroup bool        // 是否为群聊
	count   int         // 该用户的订阅数量
	channel chan string // 消息载体
}

type ReceiverManager struct {
	receivers map[string]*Receiver
	mutex     sync.RWMutex
}

func NewReceiverManager() *ReceiverManager {
	return &ReceiverManager{receivers: make(map[string]*Receiver)}
}

// 监听消息，如接收到消息则通过messager发送
func (r *Receiver) CheckMessage(bot Messager) {
	var msg string
	for {
		msg = <-r.channel
		go bot.SendMessage(r.userID, msg, r.isGroup)
	}
}

func (r *ReceiverManager) RegisterReceiver(userID int64, isGroup bool) *Receiver {
	channel := make(chan string, 5)
	receiver := &Receiver{
		userID:  userID,
		channel: channel,
		isGroup: isGroup,
	}
	key := fmt.Sprintf("%v%v", userID, isGroup)
	r.mutex.Lock()
	r.receivers[key] = receiver
	r.mutex.Unlock()
	return receiver
}

func (r *ReceiverManager) GetReceiver(userID int64, isGroup bool) (*Receiver, bool) {
	key := fmt.Sprintf("%v%v", userID, isGroup)
	r.mutex.RLock()
	receiver, ok := r.receivers[key]
	r.mutex.RUnlock()
	return receiver, ok
}
