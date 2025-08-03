package mq

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"osone/utils"
	"sync"
)

// MQ representa o servidor NATS básico com suporte a JSON
type MQ struct {
	subs_fun     map[string]map[string][]func(message, topic string)
	subscribers  map[string]map[string][]string
	services     map[string]map[string]string
	mu           sync.RWMutex
	MQTT         *utils.MQTTbroker
	reqs         map[string]chan Message
	clients      map[string]Client
	services_fun map[string]map[string]func(data string, replay func(err, data string))
	listener     net.Listener
}

// NewMQ cria uma nova instância do MQ
func NewMQ() *MQ {
	return &MQ{
		MQTT:         &utils.MQTTbroker{},
		subs_fun:     make(map[string]map[string][]func(message string, topic string)),
		subscribers:  make(map[string]map[string][]string),
		clients:      make(map[string]Client),
		services:     make(map[string]map[string]string),
		reqs:         make(map[string]chan Message),
		services_fun: make(map[string]map[string]func(data string, replay func(err string, data string))),
	}
}
func (mc *MQ) send(clientId string, msg Message) error {
	if clientId == "mybroker" {
		switch msg.CMD {
		case "PUB":
			go mc.onSubscribe(msg)
		case "REQ":
			go mc.onRequest(msg)
		case "RES":
			go mc.onResponse(msg)
		}
		return nil
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("erro ao codificar mensagem: %v", err)
	}
	_, err = fmt.Fprintf(mc.clients[clientId].cnn, "%s\n", data)
	return err
}

// Start inicia o servidor NATS na porta especificada
func (mq *MQ) Start(port string) error {
	var err error
	mq.listener, err = net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	for {
		conn, err := mq.listener.Accept()
		if err != nil {
			log.Println("Erro ao aceitar conexão:", err)
			continue
		}

		go mq.handleConnection(conn)
	}
}

// Stop encerra o servidor NATS
func (mq *MQ) Stop() {
	if mq.listener != nil {
		mq.listener.Close()
	}
}
