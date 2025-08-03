package mq

import (
	"fmt"
	"time"
)

// Publish envia uma mensagem para todos os assinantes de um tópico
func (mq *MQ) handleRequest(clientId string, message Message) {
	mq.mu.RLock()
	defer mq.mu.RUnlock()
	virtual := message.Virtual
	if mq.services[virtual] == nil {
		mq.services[virtual] = make(map[string]string)
	}
	if serviceId, ok := mq.services[virtual][message.Topic]; ok {
		mq.send(serviceId, Message{
			CMD:     "REQ",
			FromID:  clientId,
			Topic:   message.Topic,
			Payload: message.Payload,
			ReqID:   message.ReqID,
			Virtual: message.Virtual,
		})

	}
}

func (mq *MQ) handleResponse(clientId string, message Message) {
	mq.mu.RLock()
	defer mq.mu.RUnlock()

	mq.send(message.FromID, Message{
		CMD:     "RES",
		FromID:  clientId,
		Topic:   message.Topic,
		Payload: message.Payload,
		ReqID:   message.ReqID,
		Virtual: message.Virtual,
	})

}

func (mq *MQ) Request(virtual, name, payload string, timeout time.Duration) (string, error) {

	reqId := fmt.Sprintf("%d", time.Now().UnixNano())
	msg := Message{
		CMD:     "REQ",
		Topic:   name,
		Payload: payload,
		ReqID:   reqId,
		Virtual: virtual,
	}

	mq.handleRequest("mybroker", msg)
	// Cria o canal de resposta
	respChan := make(chan Message, 1)

	mq.mu.Lock()
	mq.reqs[reqId] = respChan
	mq.mu.Unlock()

	// Garante que o canal seja removido após a conclusão
	defer func() {
		mq.mu.Lock()
		delete(mq.reqs, reqId)
		mq.mu.Unlock()
	}()

	// Configura o timeout
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case resp := <-respChan:
		timer.Stop()
		delete(mq.reqs, reqId)
		if resp.Payload_err != "" {
			return "", fmt.Errorf("%v", resp.Payload_err)
		}
		return resp.Payload, nil
	case <-timer.C:
		timer.Stop()
		delete(mq.reqs, reqId)
		return "", fmt.Errorf("timeout após %v", timeout)
	}
}
func (mq *MQ) onRequest(message Message) {
	virtual := message.Virtual
	if mq.services_fun[virtual] == nil {
		mq.services_fun[virtual] = make(map[string]func(data string, replay func(err string, data string)))
	}
	if funService, ok := mq.services_fun[virtual][message.Topic]; ok {
		go funService(message.Payload, func(err, data string) {
			mq.handleResponse("mybroker", Message{
				CMD:         "RES",
				FromID:      message.FromID,
				Payload:     data,
				Payload_err: err,
				Topic:       message.Topic,
				ReqID:       message.ReqID,
				Virtual:     message.Virtual,
			})
		})
	}
}
func (mq *MQ) onResponse(message Message) {
	if ch, ok := mq.reqs[message.ReqID]; ok {
		ch <- message
	}
}

/*
	if ch, ok := mc.reqs[msg.ReqID]; ok {
				ch <- msg
			}
*/
