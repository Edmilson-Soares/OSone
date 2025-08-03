package utils

import "time"

type MQbroker struct {
	Service   func(virtual, name string, cb func(data string, replay func(err, data string))) error
	Request   func(virtual, name, payload string, timeout time.Duration) (string, error)
	Publish   func(virtual, topic, payload string) error
	Subscribe func(virtual, topic string, cb func(message, topic string)) error
}
