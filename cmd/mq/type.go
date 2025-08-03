package mq

import (
	"net"
	"regexp"
	"strings"
)

// Message estrutura a mensagem JSON
type Message struct {
	CMD         string `json:"cmd"`
	Topic       string `json:"topic"`
	Payload     string `json:"payload"`
	Payload_err string `json:"payload_err,omitempty"`
	ReqID       string `json:"reqId,omitempty"`
	Virtual     string `json:"virual,omitempty"`
	FromID      string `json:"fromId,omitempty"`
	Topic_      string `json:"topic_,omitempty"` // Alternativa para topic
}

type Client struct {
	ID      string   `json:"id"`
	cnn     net.Conn `json:"-"`
	Virtual string   `json:"virual,omitempty"`
}

func replaceWildcards(s string) string {
	return regexp.MustCompile(`\\\.\\\*`).ReplaceAllString(s, ".*")
}

func RegexpString(text string) (*regexp.Regexp, error) {
	regexPattern := "^" + regexp.QuoteMeta(text) + "$"
	regexPattern = replaceWildcards(regexPattern)
	return regexp.Compile(regexPattern)

}

func mqttToRegexpPattern(topic string) string {
	escaped := regexp.QuoteMeta(topic)
	escaped = strings.ReplaceAll(escaped, "\\+", "[^/]+")

	if strings.HasSuffix(escaped, "/#") {
		escaped = strings.TrimSuffix(escaped, "/#") + "(/.*)?"
	} else if escaped == "#" {
		escaped = ".*"
	}

	return "^" + escaped + "$"
}
func RegexpStringMQtt(topic string) (*regexp.Regexp, error) {
	pattern := mqttToRegexpPattern(topic)
	return regexp.Compile(pattern)
}
