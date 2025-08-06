package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	client "osone/go"
	"syscall"
	"time"
)

func main() {

	serverAddr := "mq://root:chave-secreta-32-bytes-123456789@37.27.39.202:4052"
	mq := client.NewMQ()

	err := mq.Connect(serverAddr)
	if err != nil {
		log.Fatalf("Erro ao conectar: %v", err)
	}
	defer mq.Disconnect()

	fmt.Printf("Conectado ao servidor em %s\n", serverAddr)
	mq.Service("test.create", func(data string, replay func(err string, data string)) {
		fmt.Println(data)
		res := map[string]string{"test": "dfdfffr"}
		bt, _ := json.Marshal(res)
		replay("", string(bt))
	})
	mq.Subscribe("app", func(message, topic string) {
		fmt.Println(message, topic, "ddd")
	})

	mq.Publish("app", "test")

	res, err := mq.Request("test.create", "ddddd", 4*time.Second)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(res, "d555555")
	// Capturar Ctrl+C para sair
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	fmt.Println("\nEncerrando cliente...")
}
