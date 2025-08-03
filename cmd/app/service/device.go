package service

import (
	"encoding/json"
	"osone/utils"

	"github.com/google/uuid"
	"github.com/zishang520/engine.io/v2/errors"
)

type AuthPermission struct {
	Subscribers []string `json:"subscribers"`
	Publichers  []string `json:"publishers"`
}

type Device struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Space       string         `json:"space"`
	Code        string         `json:"code"`
	Password    string         `json:"passworrd"`
	Permissions AuthPermission `json:"permissions"`
}

func (s *Service) AddDevice(data utils.Device) (string, error) {

	if data.ID == "" {
		data.ID = uuid.New().String()
	}

	virtual, err := s.repos.GetVirtual(data.VirtualId)
	if err != nil {
		return "", err
	}

	if data.Config == nil {
		data.Config = map[string]interface{}{}
	}
	if data.Location == nil {
		data.Location = map[string]string{}
	}
	if data.Network == nil {
		data.Network = map[string]string{}
	}
	data.Auth = map[string]string{}

	mqtt, _ := json.Marshal(utils.Auth{
		ID:        data.ID,
		Username:  data.Code,
		Password:  uuid.New().String(),
		Virtual:   "/" + virtual.Name,
		VirtualId: data.VirtualId,
		Permissions: utils.AuthPermission{
			Subscribers: []string{"/" + virtual.Name + "/" + data.Code + "/#"},
			Publichers:  []string{"/" + virtual.Name + "/" + data.Code + "/#"},
		},
	})
	mqttAuth, _ := utils.Encrypt(string(mqtt))
	data.Auth["mqtt.info"] = string(mqtt)
	data.Auth["mqtt"] = mqttAuth
	data.Auth["mq"] = mqttAuth

	id, err := s.repos.AddDevice(data)
	if err != nil {
		return "", err
	}
	return id, nil
}
func (s *Service) AuthDevice(id string) (string, error) {
	device, err := s.repos.GetDevice(id)
	if err != nil {
		return "", err
	}
	virtual, err := s.repos.GetVirtual(device.VirtualId)
	if err != nil {
		return "", err
	}

	device.Auth = map[string]string{}

	mqtt, _ := json.Marshal(utils.Auth{
		ID:        device.ID,
		Username:  device.Code,
		Password:  uuid.New().String(),
		Virtual:   "/" + virtual.Name,
		VirtualId: device.VirtualId,
		Permissions: utils.AuthPermission{
			Subscribers: []string{"/" + virtual.Name + "/" + device.Code + "/#"},
			Publichers:  []string{"/" + virtual.Name + "/" + device.Code + "/#"},
		},
	})
	mqttAuth, _ := utils.Encrypt(string(mqtt))
	device.Auth["mqtt.info"] = string(mqtt)
	device.Auth["mqtt"] = mqttAuth
	device.Auth["mq"] = mqttAuth

	err = s.repos.UpdateDevice(device)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (s *Service) EditDevice(data utils.Device) (string, error) {

	device, err := s.repos.GetDevice(data.ID)
	if err != nil {
		return "", err
	}
	device.Desc = data.Desc
	device.Name = data.Name
	device.Code = data.Code
	err = s.repos.UpdateDevice(device)
	if err != nil {
		return "", err
	}
	return device.ID, nil
}

func (s *Service) ConfigDevice(data utils.Device) (string, error) {
	device, err := s.repos.GetDevice(data.ID)
	if err != nil {
		return "", err
	}
	device.Config = data.Config
	err = s.repos.UpdateDevice(device)
	if err != nil {
		return "", err
	}
	return device.ID, nil
}

func (s *Service) NetworkDevice(data utils.Device) (string, error) {
	device, err := s.repos.GetDevice(data.ID)
	if err != nil {
		return "", err
	}
	device.Network = data.Network
	err = s.repos.UpdateDevice(device)
	if err != nil {
		return "", err
	}
	return device.ID, nil
}
func (s *Service) LocationDevice(data utils.Device) (string, error) {
	device, err := s.repos.GetDevice(data.ID)
	if err != nil {
		return "", err
	}
	device.Location = data.Location
	err = s.repos.UpdateDevice(device)
	if err != nil {
		return "", err
	}
	return device.ID, nil
}

func (s *Service) DeleteDevice(id string) (string, error) {
	device, err := s.repos.GetDevice(id)
	if err != nil {
		return "", err
	}
	err = s.repos.DeleteDevice(id)
	if err != nil {
		return "", err
	}
	return device.ID, nil
}

func (s *Service) GetDevice(id string) (utils.Device, error) {
	device, err := s.repos.GetDevice(id)
	if err != nil {
		return utils.Device{}, err
	}
	output := utils.Device{
		ID:        device.ID,
		Name:      device.Name,
		Icon:      device.Icon,
		Desc:      device.Desc,
		Code:      device.Code,
		VirtualId: device.VirtualId,
		Auth:      device.Auth,
		Location:  device.Location,
		Config:    device.Config,
		Network:   device.Network,
	}

	return output, nil
}

func (s *Service) GetCodeDevice(code string) (utils.Device, error) {
	device, err := s.repos.GetCodeDevice(code)
	if err != nil {
		return utils.Device{}, err
	}
	output := utils.Device{
		ID:        device.ID,
		Name:      device.Name,
		Icon:      device.Icon,
		Desc:      device.Desc,
		Code:      device.Code,
		VirtualId: device.VirtualId,
		Auth:      device.Auth,
		Location:  device.Location,
		Config:    device.Config,
		Network:   device.Network,
	}

	return output, nil
}

type DeviceMqtt struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Virtual     string         `json:"virtual"`
	Code        string         `json:"code"`
	Password    string         `json:"passworrd"`
	Permissions AuthPermission `json:"permissions"`
}

func (s *Service) GetLoginDevice(auth map[string]string) (any, error) {
	device, err := s.repos.GetCodeDevice(auth["code"])
	if err != nil {
		return nil, err
	}

	Auth := device.Auth["mqtt.info"]

	mqtt := DeviceMqtt{}
	err = json.Unmarshal([]byte(Auth), &mqtt)
	if err != nil {
		return nil, err
	}
	mqtt.Code = device.Code
	mqtt.Name = device.Name
	if mqtt.Password != auth["password"] {
		return nil, errors.New(" wrong password")
	}
	return mqtt, nil
}
