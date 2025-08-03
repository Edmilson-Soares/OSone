package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"tcp/cmd/client"
	"time"
)

func main() {

	serverAddr := "localhost:4222"
	mq := client.NewMQ(serverAddr)

	err := mq.Connect()
	if err != nil {
		log.Fatalf("Erro ao conectar: %v", err)
	}
	defer mq.Disconnect()

	fmt.Printf("Conectado ao servidor NATS em %s\n", serverAddr)

	// Exemplo de uso interativo
	go func() {
		// Subscrever em "exemplo"
		ch, err := mq.Subscribe("exemplo")
		if err != nil {
			log.Printf("Erro ao subscrever: %v", err)
			return
		}

		for msg := range ch {
			fmt.Printf("Mensagem recebida no tópico '%s': %s\n", msg.Topic, msg.Payload)
		}
	}()

	// Publicar uma mensagem a cada 2 segundos
	go func() {
		i := 0
		for {
			select {
			case <-mq.Quit:
				return
			case <-time.After(2 * time.Second):
				msg := fmt.Sprintf("Mensagem de exemplo %d", i)
				err := mq.Publish("exemplo", msg)
				if err != nil {
					log.Printf("Erro ao publicar: %v", err)
				} else {
					fmt.Printf("Mensagem publicada: %s\n", msg)
				}
				i++
			}
		}
	}()

	// Capturar Ctrl+C para sair
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	fmt.Println("\nEncerrando cliente...")
}
