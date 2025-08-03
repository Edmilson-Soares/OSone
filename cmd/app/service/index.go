package service

import "osone/cmd/app/repos"

type Service struct {
	repos *repos.Driver
}

func New(repos *repos.Driver) *Service {
	return &Service{repos: repos}
}
