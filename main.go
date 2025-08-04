package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"osone/cmd/mq"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Erro ao carregar arquivo .env")
	}
	bk := mq.NewMQ()

	port := os.Getenv("PORT")
	fmt.Println(" broker started in port " + port)
	bk.Start(port)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	fmt.Println("\nEncerrando OSone...")
}
