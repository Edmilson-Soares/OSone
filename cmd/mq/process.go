package mq

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
)

// handleConnection manipula uma conexão de cliente
func (mq *MQ) handleProcess(clientId string) {
	fmt.Println("client connected", clientId)
	virtual := mq.clients[clientId].Virtual
	scanner := bufio.NewScanner(mq.clients[clientId].cnn)
	for scanner.Scan() {
		rawMsg := scanner.Text()

		var msg Message
		err := json.Unmarshal([]byte(rawMsg), &msg)
		if err != nil {
			log.Printf("Erro ao decodificar JSON: %v\n", err)
			fmt.Fprintf(mq.clients[clientId].cnn, `{"error":"invalid JSON"}`+"\n")
			continue
		}
		msg.Virtual = virtual

		switch msg.CMD {
		case "SUB":
			go mq.handleSubscribe(clientId, virtual, msg.Topic)
		case "PUB":
			go mq.handlePublish(virtual, msg.Topic, msg.Payload)
		case "SER":
			go mq.handleService(clientId, virtual, msg.Topic)
		case "REQ":
			go mq.handleRequest(clientId, msg)
		case "RES":
			go mq.handleResponse(clientId, msg)
		}

		// Se topic_ estiver definido, usa como fallback para topic

	}
}
