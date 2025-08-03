package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"osone/cmd/client"
	"syscall"
	"time"
)

func main() {

	serverAddr := "mq://root:admin@localhost:4222"
	mq := client.NewMQ()

	err := mq.Connect(serverAddr)
	if err != nil {
		log.Fatalf("Erro ao conectar: %v", err)
	}
	defer mq.Disconnect()

	mq.Service("exemplo", func(data string, replay func(err string, data string)) {
		fmt.Println(data, "dddd")
		replay("", "ddddddddddddddddddd")
	})

	mq.Subscribe("test.*.test", func(message, topic string) {
		fmt.Println(message, topic, "ddd")
	})

	mq.Subscribe("test.ok.test", func(message, topic string) {
		fmt.Println(message, topic)
	})
	mq.Publish("test.ok.test", "test")

	res, err := mq.Request("test", "ddddd", 4*time.Second)
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
