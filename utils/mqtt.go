package utils

type MQTTbroker struct {
	Publish   func(topic, payload string)
	Subscribe func(topic string, cb func(message, topic string), subId int)
}
