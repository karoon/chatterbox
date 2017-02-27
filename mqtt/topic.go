package mqtt

import (
	"sync"
)

type Topic struct {
	Content         string
	RetainedMessage *MqttMessage
}

var globalTopicss = make(map[string]*Topic)
var globalTopicssLockk = new(sync.Mutex)

func CreateTopic(content string) *Topic {
	topic := new(Topic)
	topic.Content = content
	topic.RetainedMessage = nil
	return topic
}
