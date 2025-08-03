package mqtt

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"osone/utils"
	"syscall"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/mochi-mqtt/server/v2/packets"
)

type Broker struct {
	server  *mqtt.Server
	kv      *KV
	event   *EventEmitter
	mq      *utils.MQbroker
	clients map[string]Device
}

func (bk *Broker) Start() {

	bk.kv = bk.authKV()

	bk.event.On("connected", func(client interface{}) {
		device := client.(Device)
		str, _ := json.Marshal(device)
		bk.mq.Publish(device.Virtual, "mqtt.connected", string(str))
		//str, _ := json.Marshal(device)
		//bk.mq.Publish("mqtt.connected", string(str))
		//bk.mq.Publish(device.Space+"/mqtt.connected", string(str))
		//fmt.Println("connected", device)
	})
	bk.event.On("disconnected", func(client interface{}) {
		device := client.(Device)
		str, _ := json.Marshal(device)
		bk.mq.Publish(device.Virtual, "mqtt.disconnected", string(str))
		//fmt.Println("disconnected", device)
		//str, _ := json.Marshal(device)
		//bk.mq.Publish(device.Space+"/mqtt.disconnected", string(str))
		//bk.mq.Publish("mqtt.disconnected", string(str))

	})
	// Create signals channel to run server until interrupted
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		done <- true
	}()

	// Create the new MQTT Server.
	bk.server = mqtt.New(&mqtt.Options{
		SysTopicResendInterval: 10,
		InlineClient:           true,
	})

	// Allow all connections.
	_ = bk.server.AddHook(new(AllowHook), bk.authHook())

	// Create a TCP listener on a standard port.
	tcp := listeners.NewTCP(listeners.Config{ID: "t1", Address: ":" + os.Getenv("MQTT_PORT")})
	err := bk.server.AddListener(tcp)
	if err != nil {
		log.Fatal(err)
	}
	go bk.run()
	go func() {
		err := bk.server.Serve()
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Run server until interrupted
	<-done

	// Cleanup
}
func (bk *Broker) run() {

	/*
		bk.nats.AddService("mqtt.addDevice", func(data string, reply func(res string)) {
			var device Device
			err := json.Unmarshal([]byte(data), &device)
			if err != nil {
				output := fmt.Sprintf(`{"error":"%s"}`, err.Error())
				reply(output)
				return
			}

			err = bk.kv.addDevice(device)
			if err != nil {
				output := fmt.Sprintf(`{"error":"%s"}`, err.Error())
				reply(output)
				return
			}
			reply(`{"ok":true}`)
		})

		bk.nats.AddService("mqtt.delDevice", func(data string, reply func(res string)) {

			err := bk.kv.delDevice(data)
			if err != nil {
				output := fmt.Sprintf(`{"error":"%s"}`, err.Error())
				reply(output)
				return
			}

			reply(`{"ok":true}`)
		})
		bk.nats.AddService("mqtt.getDevice", func(data string, reply func(res string)) {
			device, err := bk.kv.getDevice(data)
			if err != nil {
				output := fmt.Sprintf(`{"error":"%s"}`, err.Error())
				reply(output)
				return
			}
			str, _ := json.Marshal(device)
			reply(string(str))
		})

		bk.nats.AddService("mqtt.publish", func(data string, reply func(res string)) {
			dataMqtt := map[string]string{}
			err := json.Unmarshal([]byte(data), &dataMqtt)
			if err != nil {
				output := fmt.Sprintf(`{"error":"%s"}`, err.Error())
				reply(output)
				return
			}
			err = bk.server.Publish(dataMqtt["topic"], []byte(dataMqtt["playload"]), false, 0)
			if err != nil {
				output := fmt.Sprintf(`{"error":"%s"}`, err.Error())
				reply(output)
				return
			}

			reply(`{"ok":true}`)
		})

		bk.nats.Sub("mqtt.publish", func(data string) {

			dataMqtt := map[string]string{}
			err := json.Unmarshal([]byte(data), &dataMqtt)
			if err != nil {
				return
			}
			err = bk.server.Publish(dataMqtt["topic"], []byte(dataMqtt["playload"]), false, 0)
			if err != nil {
				return
			}

		})
	*/

	//err := server.Publish("direct/publish", []byte("packet scheduled message"), false, 0)
	/*

		bk.kv.addDevice(Device{
			ID:   "HHGJDHDJDKDJHDJBGJDBJUHBDJBDJBDBJDBJDBJBDJDB",
			Code: "john_doe",
			Permissions: AuthPermission{
				Subscribers: []string{"teste/#"},
				Publichers:  []string{"teste/#"},
			},
		})
	*/

	/*

			bk.mq.Subscribe("mqtt.publish", func(msg mq.MQData) {
				dataMqtt := map[string]string{}
				err := json.Unmarshal([]byte(msg.Payload), &dataMqtt)
				if err != nil {
					return
				}
				err = bk.server.Publish(dataMqtt["topic"], []byte(dataMqtt["playload"]), false, 0)
				if err != nil {
					return
				}
			})


				bk.mq.Service("mqtt.getDeviceAll", func(msg mq.MQData, replay func(err string, payload string)) {
			devices := []Device{}
			for _, device := range bk.clients {

				if msg.Payload != "" {
					if device.Space == msg.Payload {
						devices = append(devices, Device{
							ID:    device.ID,
							Code:  device.Code,
							Space: device.Space,
						})
					}

				} else {
					devices = append(devices, Device{
						ID:    device.ID,
						Code:  device.Code,
						Space: device.Space,
					})
				}

			}

			str, _ := json.Marshal(devices)
			replay("", string(str))
		})

	*/

}

// func(topic string, cb func(message, topic string)) error
func (bk *Broker) Subscribe(topic string, cb func(message, topic string), subId int) {

	callbackFn := func(cl *mqtt.Client, sub packets.Subscription, pk packets.Packet) {
		cb(pk.TopicName, string(pk.Payload))
	}

	err := bk.server.Subscribe(topic, subId, callbackFn)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (bk *Broker) Publish(topic, message string) {
	err := bk.server.Publish(topic, []byte(message), false, 0)
	if err != nil {
		fmt.Println(err)
		return
	}
}
func NewBroker(mq *utils.MQbroker) *Broker {
	//nats_url := os.Getenv("NATS_URL")
	return &Broker{
		clients: map[string]Device{},
		event:   NewEventEmitter(),
		mq:      mq,
	}
}
