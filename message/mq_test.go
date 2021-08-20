package message

import (
	"reflect"
	"sort"
	"testing"
)

// 用于获取message的返回值
var message = make(chan string, 1)

type bot struct{}

func (b *bot) SendMessage(account int64, msg string, isGroup bool) {
	message <- msg
}

func TestMessageQueue_Close(t *testing.T) {
	var (
		mq = New()
	)
	if !mq.checkStatus() {
		t.Error("消息队列创建出现错误")
	}
	mq.Close()
	if mq.checkStatus() {
		t.Error("消息队列关闭出现错误")
	}
}

func TestMessageQueue_Publish(t *testing.T) {
	var (
		mq      = New()
		testBot = &bot{}
	)
	tests := []struct {
		name string
		args []string
	}{
		{"单条消息", []string{"test"}},
		{"多条消息", []string{"this", "is", "a", "message"}},
	}
	mq.Subscribe(testBot, "topic", 0, false)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				messages []string
				count    int
			)
			for _, arg := range tt.args {
				go mq.Publish("topic", arg)
			}
			for msg := range message {
				messages = append(messages, msg)
				count++
				if count == len(tt.args) {
					break
				}
			}
			sort.Strings(tt.args)
			sort.Strings(messages)
			if !reflect.DeepEqual(messages, tt.args) {
				t.Errorf("error receiving message,want %v,got %v", tt.args, messages)
			}
		})
	}
	mq.Close()
}
