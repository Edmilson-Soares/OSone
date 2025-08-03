package MQapp

import (
	"encoding/json"
	"fmt"
	"osone/utils"
)

func (s MQService) device() {

	s.mq.Service("/", "osone.addDevice", func(data string, replay func(err string, data string)) {
		input := utils.Device{}
		err := json.Unmarshal([]byte(data), &input)
		if err != nil {
			replay(err.Error(), "")
			return
		}
		id, err := s.service.AddDevice(input)
		if err != nil {
			replay(err.Error(), "")
			return
		}
		res := fmt.Sprintf(`{ "id": "%s" }`, id)
		replay("", res)
	})
	s.mq.Service("/", "osone.editDevice", func(data string, replay func(err string, data string)) {
		input := utils.Device{}
		err := json.Unmarshal([]byte(data), &input)
		if err != nil {
			replay(err.Error(), "")
			return
		}
		id, err := s.service.EditDevice(input)
		if err != nil {
			replay(err.Error(), "")
			return
		}
		res := fmt.Sprintf(`{ "id": "%s" }`, id)
		replay("", res)
	})
	s.mq.Service("/", "osone.networkDevice", func(data string, replay func(err string, data string)) {
		input := map[string]string{}
		err := json.Unmarshal([]byte(data), &input)
		if err != nil {
			replay(err.Error(), "")
			return
		}

		input_ := utils.Device{
			ID:      input["id"],
			Network: make(map[string]string),
		}
		delete(input, "id")
		input_.Network = input
		id, err := s.service.NetworkDevice(input_)
		if err != nil {
			replay(err.Error(), "")
			return
		}
		res := fmt.Sprintf(`{ "id": "%s" }`, id)
		replay("", res)
	})

	s.mq.Service("/", "osone.locationDevice", func(data string, replay func(err string, data string)) {
		input := map[string]string{}
		err := json.Unmarshal([]byte(data), &input)
		if err != nil {
			replay(err.Error(), "")
			return
		}

		input_ := utils.Device{
			ID:       input["id"],
			Location: make(map[string]string),
		}
		delete(input, "id")
		input_.Location = input
		id, err := s.service.LocationDevice(input_)
		if err != nil {
			replay(err.Error(), "")
			return
		}
		res := fmt.Sprintf(`{ "id": "%s" }`, id)
		replay("", res)
	})
	s.mq.Service("/", "osone.configDevice", func(data string, replay func(err string, data string)) {
		input := map[string]interface{}{}
		err := json.Unmarshal([]byte(data), &input)
		if err != nil {
			replay(err.Error(), "")
			return
		}

		input_ := utils.Device{
			ID:     input["id"].(string),
			Config: make(map[string]interface{}),
		}
		delete(input, "id")
		input_.Config = input
		id, err := s.service.ConfigDevice(input_)
		if err != nil {
			replay(err.Error(), "")
			return
		}
		res := fmt.Sprintf(`{ "id": "%s" }`, id)
		replay("", res)
	})
	s.mq.Service("/", "osone.delDevice", func(data string, replay func(err string, data string)) {
		input := map[string]string{}
		err := json.Unmarshal([]byte(data), &input)
		if err != nil {
			replay(err.Error(), "")
			return
		}
		id, err := s.service.DeleteDevice(input["id"])
		if err != nil {
			replay(err.Error(), "")
			return
		}
		res := fmt.Sprintf(`{ "id": "%s" }`, id)
		replay("", res)
	})

	s.mq.Service("/", "osone.authDevice", func(data string, replay func(err string, data string)) {
		input := map[string]string{}
		err := json.Unmarshal([]byte(data), &input)
		if err != nil {
			replay(err.Error(), "")
			return
		}
		id, err := s.service.AuthDevice(input["id"])
		if err != nil {
			replay(err.Error(), "")
			return
		}
		res := fmt.Sprintf(`{ "id": "%s" }`, id)
		replay("", res)
	})

	s.mq.Service("/", "osone.device.mqtt", func(data string, replay func(err string, data string)) {

		dataMqtt := map[string]string{}
		err := json.Unmarshal([]byte(data), &dataMqtt)
		if err != nil {
			return
		}

		device, err := s.service.GetLoginDevice(dataMqtt)
		if err != nil {
			replay(err.Error(), "")
			return
		}

		str, err := json.Marshal(device)
		if err != nil {
			replay(err.Error(), "")
			return
		}
		replay("", string(str))
	})

}
