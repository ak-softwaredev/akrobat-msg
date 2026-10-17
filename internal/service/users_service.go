package service

import "github.com/ak-softwaredev/akrobat-msg/internal/repository"

type UsersService interface {
}

type usersService struct {
	users repository.UsersRepository
}

func NewUsersService(users repository.UsersRepository) UsersService {
	return &usersService{
		users: users,
	}
}
