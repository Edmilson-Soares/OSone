package mq

import (
	"fmt"
	"net"

	"github.com/google/uuid"
)

// handleConnection manipula uma conexão de cliente
func (mq *MQ) handleConnection(conn net.Conn) {
	clientId := uuid.New().String()
	defer func() {
		conn.Close()
		go mq.handleClose(clientId)
	}()

	mq.clients[clientId] = Client{
		ID:  clientId,
		cnn: conn,
	}

	if err := mq.handleAuth(clientId); err != nil {

	} else {
		mq.handleProcess(clientId)
	}

}

func (mq *MQ) handleClose(clientId string) {

	virtual := mq.clients[clientId].Virtual
	mq.clients[clientId].cnn.Close()
	delete(mq.clients, clientId)
	for topic, subscribers := range mq.subscribers[virtual] {
		// Filtra a lista de subscribers, removendo o clientId
		newSubscribers := make([]string, 0, len(subscribers))
		for _, sub := range subscribers {
			if sub != clientId {
				newSubscribers = append(newSubscribers, sub)
			}
		}
		// Atualiza a lista de subscribers para o tópico
		mq.subscribers[virtual][topic] = newSubscribers
	}
	for topic, _ := range mq.services {
		if mq.services[virtual][topic] == clientId {
			delete(mq.services, clientId)
		}
	}

	fmt.Println("client disconnected", clientId)

}
