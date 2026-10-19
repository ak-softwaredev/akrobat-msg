package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/ak-softwaredev/akrobat-msg/internal/model"
	"github.com/ak-softwaredev/akrobat-msg/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUsernameBadLength = errors.New("username length off-limits")
	ErrPasswordBadLength = errors.New("password length off-limits")
	ErrInvalidUserField  = errors.New("invalid user field")
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

func (s *usersService) GetByID(
	ctx context.Context,
	id int64,
) (*model.User, error) {
	u, err := s.users.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("users get by id: %w", err)
	}

	return u, nil
}

func (s *usersService) Update(
	ctx context.Context,
	id int64,
	field string,
	value string,
) (*model.User, error) {
	var u *model.User
	var err error

	switch field {
	case "bio":
		u, err = s.users.UpdateBio(ctx, id, value)
	case "username":
		validationErr := validateAuthInput(&value, nil)
		if validationErr != nil {
			return nil, validationErr
		}

		u, err = s.users.UpdateUsername(ctx, id, value)
	case "password":
		validationErr := validateAuthInput(nil, &value)
		if validationErr != nil {
			return nil, validationErr
		}

		passwordHash, err := bcrypt.GenerateFromPassword(
			[]byte(value),
			bcrypt.DefaultCost,
		)
		if err != nil {
			return nil, fmt.Errorf("auth register: %w", err)
		}

		u, err = s.users.UpdatePassword(ctx, id, string(passwordHash))
	default:
		return nil, ErrInvalidUserField
	}
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
	}

	return u, nil
}

func (s *usersService) Delete(
	ctx context.Context,
	id int64,
) error {
	err := s.users.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	return nil
}
