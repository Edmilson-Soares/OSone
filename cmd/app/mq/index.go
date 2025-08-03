package MQapp

import (
	"osone/cmd/app/service"
	"osone/utils"
)

type MQService struct {
	service *service.Service
	mq      *utils.MQbroker
}

func New(service *service.Service, mq *utils.MQbroker) *MQService {
	return &MQService{service: service, mq: mq}
}

func (s *MQService) Run() {
	s.virtual()
	s.device()
}
