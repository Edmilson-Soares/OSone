package MQapp

import (
	"encoding/json"
	"fmt"
	"osone/utils"
)

func (s *MQService) virtual() {

	s.mq.Service("/", "osone.addVirtual", func(data string, replay func(err string, data string)) {
		input := utils.Virtual{}
		err := json.Unmarshal([]byte(data), &input)
		if err != nil {
			replay(err.Error(), "")
			return
		}
		id, err := s.service.AddVirtual(input)
		if err != nil {
			replay(err.Error(), "")
			return
		}
		res := fmt.Sprintf(`{ "id": "%s" }`, id)
		replay("", res)
	})

	s.mq.Service("/", "osone.editVirtual", func(data string, replay func(err string, data string)) {
		input := utils.Virtual{}
		err := json.Unmarshal([]byte(data), &input)
		if err != nil {
			replay(err.Error(), "")
			return
		}
		id, err := s.service.EditVirtual(input)
		if err != nil {
			replay(err.Error(), "")
			return
		}
		res := fmt.Sprintf(`{ "id": "%s" }`, id)
		replay("", res)
	})
	s.mq.Service("/", "osone.authVirtual", func(data string, replay func(err string, data string)) {
		id := data
		_, err := s.service.AuthDevice(id)
		if err != nil {
			replay(err.Error(), "")
			return
		}
		res := fmt.Sprintf(`{ "id": "%s" }`, id)
		replay("", res)
	})

	s.mq.Service("/", "osone.delVirtual", func(data string, replay func(err string, data string)) {
		input := utils.Virtual{}
		err := json.Unmarshal([]byte(data), &input)
		if err != nil {
			replay(err.Error(), "")
			return
		}
		id, err := s.service.DeleteVirtual(input)
		if err != nil {
			replay(err.Error(), "")
			return
		}
		res := fmt.Sprintf(`{ "id": "%s" }`, id)
		replay("", res)
	})

	s.mq.Service("/", "osone.getVirtual", func(data string, replay func(err string, data string)) {

		id := data
		res, err := s.service.GetVirtual(id)
		if err != nil {
			replay(err.Error(), "")
			return
		}

		by, err := json.Marshal(res)
		if err != nil {
			replay(err.Error(), "")
			return
		}
		replay("", string(by))
	})
}
