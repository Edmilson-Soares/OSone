package service

import (
	"encoding/json"
	"fmt"
	"osone/utils"

	"github.com/google/uuid"
)

func (s *Service) AddVirtual(data utils.Virtual) (string, error) {

	if data.ID == "" {
		data.ID = uuid.New().String()
	}

	mq := fmt.Sprintf(`{"id":"%s","virtual":"/%s"}`, data.ID, data.Name)
	mqAuth, err := utils.Encrypt(mq)
	if err != nil {
		return "", err
	}
	mqtt, _ := json.Marshal(utils.Auth{
		ID:        data.ID,
		Username:  data.Name,
		Password:  uuid.New().String(),
		Virtual:   "/" + data.Name,
		VirtualId: data.ID,
		Permissions: utils.AuthPermission{
			Subscribers: []string{"/" + data.Name + "/#"},
			Publichers:  []string{"/" + data.Name + "/#"},
		},
	})
	mqttAuth, _ := utils.Encrypt(string(mqtt))
	data.Auth = map[string]string{"mq": mqAuth, "mqtt": mqttAuth}
	id, err := s.repos.AddVirtual(data)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (s *Service) EditVirtual(data utils.Virtual) (string, error) {

	virtual, err := s.repos.GetVirtual(data.ID)
	if err != nil {
		return "", err
	}
	virtual.Desc = data.Desc
	virtual.Icon = data.Icon

	err = s.repos.UpdateVirtual(virtual)
	if err != nil {
		return "", err
	}
	return virtual.ID, nil
}
func (s *Service) AuthVirtual(id string) (string, error) {

	virtual, err := s.repos.GetVirtual(id)
	if err != nil {
		return "", err
	}
	mq := fmt.Sprintf(`{"id":"%s","virtual":"/%s"}`, virtual.ID, virtual.Name)
	mqAuth, err := utils.Encrypt(mq)
	if err != nil {
		return "", err
	}
	mqtt, _ := json.Marshal(utils.Auth{
		ID:        virtual.ID,
		Username:  virtual.Name,
		Password:  uuid.New().String(),
		Virtual:   "/" + virtual.Name,
		VirtualId: virtual.ID,
		Permissions: utils.AuthPermission{
			Subscribers: []string{"/" + virtual.Name + "/#"},
			Publichers:  []string{"/" + virtual.Name + "/#"},
		},
	})
	mqttAuth, _ := utils.Encrypt(string(mqtt))
	virtual.Auth = map[string]string{"mq": mqAuth, "mqtt": mqttAuth}
	err = s.repos.UpdateVirtual(virtual)
	if err != nil {
		return "", err
	}
	return virtual.ID, nil
}
func (s *Service) DeleteVirtual(data utils.Virtual) (string, error) {
	virtual, err := s.repos.GetVirtual(data.ID)
	if err != nil {
		return "", err
	}
	err = s.repos.DeleteVirtual(virtual.ID)
	if err != nil {
		return "", err
	}
	return virtual.ID, nil
}

func (s *Service) GetVirtual(id string) (utils.Virtual, error) {

	virtualRepo, err := s.repos.GetVirtualView(id)
	if err != nil {
		return utils.Virtual{}, err
	}

	virtual := utils.Virtual{
		ID:           virtualRepo.ID,
		Name:         virtualRepo.Name,
		Desc:         virtualRepo.Desc,
		Icon:         virtualRepo.Icon,
		Auth:         virtualRepo.Auth,
		EnterpriseId: virtualRepo.EnterpriseId,
		Devices:      make([]utils.Device, 0),
	}
	for _, device := range virtualRepo.Devices {
		virtual.Devices = append(virtual.Devices, utils.Device{
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
		})
	}
	return virtual, nil
}
