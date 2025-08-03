package mqtt

import (
	"encoding/json"
	"fmt"
	"osone/utils"
	"regexp"
	"strings"
	"time"
)

type HookAuth struct {
	kv  *KV
	mq  *utils.MQbroker
	add func(client Device)
	del func(id string)
	get func(id string) Device
}

func (h *HookAuth) mqttToRegexpPattern(topic string) string {
	escaped := regexp.QuoteMeta(topic)
	escaped = strings.ReplaceAll(escaped, "\\+", "[^/]+")

	if strings.HasSuffix(escaped, "/#") {
		escaped = strings.TrimSuffix(escaped, "/#") + "(/.*)?"
	} else if escaped == "#" {
		escaped = ".*"
	}

	return "^" + escaped + "$"
}

func (h *HookAuth) RegexpString(topic string) (*regexp.Regexp, error) {
	pattern := h.mqttToRegexpPattern(topic)
	return regexp.Compile(pattern)
}

func (bk HookAuth) login(id, username, password string) bool {

	if username == "apikey" {
		device := Device{}
		str, err := Decrypt(password)
		if err != nil {
			return false
		}
		err = json.Unmarshal([]byte(str), &device)
		if err != nil {
			return false
		}

		device.ID = id
		bk.add(device)
		return true

	}

	if username == "device" {

		device, err := bk.kv.getDevice(password)
		if err != nil {
			return false
		}
		device.ID = id
		bk.add(device)
		return true

	}
	if username == "app" {

		return false
	}

	device := Device{}
	input := fmt.Sprintf(`{"code":"%s","password":"%s"}`, username, password)
	str, err := bk.mq.Request("/", "osone.device.mqtt", input, 5*time.Second)
	if err != nil {
		return false
	}

	err = json.Unmarshal([]byte(str), &device)
	if err != nil {
		return false
	}
	device.ID = id
	bk.add(device)

	return true
}

func (bk HookAuth) ACLCheck(id, topic string, write bool) bool {
	auth := bk.get(id)
	ok := false

	var topicsToCheck []string
	if write {
		topicsToCheck = auth.Permissions.Publichers
	} else {
		topicsToCheck = auth.Permissions.Subscribers
	}

	for _, allowedTopic := range topicsToCheck {

		allowedTopic_ := allowedTopic

		re, err := bk.RegexpString(allowedTopic_)
		if err != nil {
			continue
		}

		if re.MatchString(topic) {
			ok = true
			break
		}
	}
	return ok
}

func (bk Broker) authHook() *HookAuth {
	return &HookAuth{kv: bk.kv, del: bk.delClient, add: bk.addClient, get: bk.getClient, mq: bk.mq}
}

func (bk Broker) addClient(client Device) {
	bk.clients[client.ID] = client
	bk.event.Emit("connected", client)

}

func (bk Broker) delClient(id string) {
	device := bk.clients[id]
	bk.event.Emit("disconnected", device)
	delete(bk.clients, id)
}

func (bk Broker) getClient(id string) Device {
	return bk.clients[id]
}
