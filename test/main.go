package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	client "osone/go"
	"syscall"
)

func main() {

	serverAddr := "mq://apikey:8rQJ0Afl1sY3NfC57b67X4sgrciuUoVWFtKO4IZQSH_4HsfkVb8yYJAXfCF_nyJTdIpG0C1_IcCfG_Jn_5fJzj2W5osg3A90nPnq3cMi2Oysj8SNU_yuNSDfVw==@localhost:4222" //"mq://root:chave-secreta-32-bytes-123456789@localhost:4222"
	mq := client.NewMQ()

	err := mq.Connect(serverAddr)
	if err != nil {
		log.Fatalf("Erro ao conectar: %v", err)
	}
	defer mq.Disconnect()

	fmt.Printf("Conectado ao servidor em %s\n", serverAddr)
	mq.Subscribe("mqtt.connected", func(message, topic string) {
		fmt.Println(message, topic, "mqtt")
	})
	mq.Subscribe("mqtt.disconnected", func(message, topic string) {
		fmt.Println(message, topic, "mqtt")
	})
	/*mq.Subscribe("test", func(message, topic string) {
		fmt.Println(message, topic, "ddd")
	})
	mq.Subscribe("mqtt::test", func(message, topic string) {
		fmt.Println(message, topic, "mqtt")
	})

	//mq.Publish("test", "test")

	mq.Publish("mqtt::test", "test")

	*/

	mq.Service("test.create", func(data string, replay func(err string, data string)) {
		fmt.Println(data)
		res := map[string]string{"test": "dfdfffr"}
		bt, _ := json.Marshal(res)
		replay("", string(bt))
	})
	// Capturar Ctrl+C para sair
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	fmt.Println("\nEncerrando cliente...")
}
