package main

import (
	"fmt"
	"log"
	"os"
	"osone/cmd/api"
	MQapp "osone/cmd/app/mq"
	"osone/cmd/app/repos"
	"osone/cmd/app/service"
	"osone/cmd/mq"
	"osone/cmd/mqtt"
	"osone/utils"
)

type OSone struct {
	mq   *mq.MQ
	mqtt *mqtt.Broker
	api  *api.Server
}

func (osone *OSone) run() {

	defer osone.mq.Stop()

	repo, err := repos.NewDB()
	if err != nil {
		log.Fatal(err)
	}

	service_ := service.New(repo)
	appmq := &utils.MQbroker{Service: osone.mq.Service, Request: osone.mq.Request, Subscribe: osone.mq.Subscribe, Publish: osone.mq.Publish}
	app := MQapp.New(service_, appmq)
	osone.mqtt = mqtt.NewBroker(appmq)

	appmqtt := &utils.MQTTbroker{Publish: osone.mqtt.Publish, Subscribe: osone.mqtt.Subscribe}
	osone.api = api.NewServer(appmq, appmqtt)
	go osone.mqtt.Start()
	go osone.api.Run(":" + os.Getenv("API_PORT"))
	go app.Run()
	go func() {
		osone.mq.MQTT = appmqtt
		port := os.Getenv("PORT")
		fmt.Println(" broker started in port " + port)
		if err := osone.mq.Start(port); err != nil {
			log.Fatal(err)
		}

	}()

}

func NewOSone() *OSone {
	sys := &OSone{
		mq: mq.NewMQ(),
	}

	return sys
}
