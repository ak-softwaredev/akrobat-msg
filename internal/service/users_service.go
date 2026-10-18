package service

import (
	"context"

	"github.com/ak-softwaredev/akrobat-msg/internal/model"
)

type UsersRepository interface {
	Register(ctx context.Context, username, passwordHash string) (*model.User, error)
	GetWithCredentialsByUsername(ctx context.Context, username string) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByID(ctx context.Context, id int64) (*model.User, error)
	UpdateUsername(ctx context.Context, id int64, username string) (*model.User, error)
	UpdatePassword(ctx context.Context, id int64, passwordHash string) (*model.User, error)
	UpdateBio(ctx context.Context, id int64, bio string) (*model.User, error)
	Delete(ctx context.Context, id int64) error
}

type usersService struct {
	users UsersRepository
}

func NewUsersService(
	users UsersRepository,
) *usersService {
	return &usersService{
		users: users,
	}
}
