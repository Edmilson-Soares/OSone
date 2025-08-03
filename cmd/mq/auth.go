package mq

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"osone/utils"
)

// handleConnection manipula uma conexão de cliente
func (mq *MQ) handleAuth(clientId string) error {
	ok := false
	id := ""
	scanner := bufio.NewScanner(mq.clients[clientId].cnn)
	virtual := ""
	for scanner.Scan() {
		rawMsg := scanner.Text()

		var msg Message
		err := json.Unmarshal([]byte(rawMsg), &msg)
		if err != nil {
			log.Printf("Erro ao decodificar JSON: %v\n", err)
			fmt.Fprintf(mq.clients[clientId].cnn, `{"error":"invalid JSON"}`+"\n")
			break
		}

		switch msg.Topic {
		case "apikey":
			str, err := utils.Decrypt(msg.Payload)
			if err != nil {
				mq.send(clientId, Message{
					CMD:         "RES",
					Topic:       "AUHT",
					Payload:     clientId,
					Payload_err: "not authorized",
					ReqID:       msg.ReqID,
				})

				break
			}

			input := map[string]string{}
			err = json.Unmarshal([]byte(str), &input)
			if err != nil {
				mq.send(clientId, Message{
					CMD:         "RES",
					Topic:       "AUHT",
					Payload:     clientId,
					Payload_err: "not authorized",
					ReqID:       msg.ReqID,
				})

				break
			}

			ok = true
			mq.send(clientId, Message{
				CMD:         "RES",
				Topic:       "AUHT",
				Payload:     clientId,
				Payload_err: "",
				ReqID:       msg.ReqID,
			})

			virtual = input["virtual"]
			id = input["id"]
			break
		case "root":
			if msg.Payload == os.Getenv("ROOT_PASSWORD") {
				ok = true
				mq.send(clientId, Message{
					CMD:         "RES",
					Topic:       "AUHT",
					Payload:     clientId,
					Payload_err: "",
					ReqID:       msg.ReqID,
				})
				id = clientId
				virtual = "/"
			} else {
				mq.send(clientId, Message{
					CMD:         "RES",
					Topic:       "AUHT",
					Payload:     clientId,
					Payload_err: "not authorized",
					ReqID:       msg.ReqID,
				})
			}
			break
		}

		break
	}

	if !ok {
		return errors.New("auth failed")
	}
	mq.clients[clientId] = Client{
		ID:      id,
		cnn:     mq.clients[clientId].cnn,
		Virtual: virtual,
	}
	return nil

}
