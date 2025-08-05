package mq

import (
	"strings"
)

// Publish envia uma mensagem para todos os assinantes de um tópico
func (mq *MQ) handlePublish(virtual, topic, message string) {
	mq.mu.RLock()
	defer mq.mu.RUnlock()

	if mq.subscribers[virtual] == nil {
		mq.subscribers[virtual] = make(map[string][]string)
	}

	if strings.HasPrefix(topic, "mqtt::") && !strings.HasPrefix(topic, "pub") {

		topic_ := strings.Split(topic, "mqtt::")[1]
		if mq.MQTT.Publish != nil {
			mq.MQTT.Publish(topic_, message)
		}

		return
	}

	topic_ := topic
	if strings.HasPrefix(topic, "pub") {
		topic_ = strings.Split(topic, "pub")[1]
	}

	for strtopic, _ := range mq.subscribers[virtual] {

		if strings.HasPrefix(topic_, "mqtt::") {
			re, err := RegexpStringMQtt(strtopic)
			if err != nil {
				continue
			}
			if re.MatchString(topic_) {
				for _, subId := range mq.subscribers[virtual][strtopic] {
					mq.send(subId, Message{
						CMD:     "PUB",
						Topic_:  topic_,
						Virtual: virtual,
						Topic:   strtopic,
						Payload: message,
					})

				}

			}
		} else {

			//topic_ := strings.Split(topic, "pubmqtt::")[1]
			re, err := RegexpString(strtopic)
			if err != nil {

				continue
			}
			if re.MatchString(topic_) {
				for _, subId := range mq.subscribers[virtual][strtopic] {
					mq.send(subId, Message{
						CMD:     "PUB",
						Topic_:  topic_,
						Virtual: virtual,
						Topic:   strtopic,
						Payload: message,
					})

				}

			}
		}
	}

}
func (mq *MQ) onSubscribe(message Message) error {
	mq.mu.RLock()
	defer mq.mu.RUnlock()
	virtual := message.Virtual
	if mq.subs_fun[virtual] == nil {
		mq.subs_fun[virtual] = make(map[string][]func(message string, topic string))
	}
	for _, funSub := range mq.subs_fun[virtual][message.Topic] {
		go funSub(message.Payload, message.Topic_)
	}
	return nil
}

// Publish envia uma mensagem para um tópico
func (mq *MQ) Publish(virtual, topic, payload string) error {
	mq.handlePublish(virtual, topic, payload)
	return nil
}
