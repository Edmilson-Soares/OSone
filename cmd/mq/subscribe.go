package mq

import (
	"strings"
)

// Subscribe adiciona um novo assinante para um tópico

func (mq *MQ) handleSubscribe(clientId, virtual, topic string) {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	if mq.subscribers[virtual] == nil {
		mq.subscribers[virtual] = make(map[string][]string)
	}

	if strings.HasPrefix(topic, "mqtt::") {
		topic_ := strings.Split(topic, "mqtt::")[1]
		if mq.MQTT.Subscribe != nil {
			mq.MQTT.Subscribe(topic_, func(topic, message string) {
				mq.Publish(virtual, "pubmqtt::"+topic_, message)
			}, 2)

		}

	}
	mq.subscribers[virtual][topic] = append(mq.subscribers[virtual][topic], clientId)
}
func (mq *MQ) Subscribe(virtual, topic string, cb func(message, topic string)) error {
	mq.handleSubscribe("mybroker", virtual, topic)
	if mq.subs_fun[virtual] == nil {
		mq.subs_fun[virtual] = make(map[string][]func(message string, topic string))
	}
	mq.subs_fun[virtual][topic] = append(mq.subs_fun[virtual][topic], cb)
	return nil
}
