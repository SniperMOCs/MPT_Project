package user

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) Create(ctx context.Context, u *User) error {
	exists, err := s.repo.ExistsByUsername(ctx, u.Username)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("username taken")
	}
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(u.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return err
	}
	u.ID = uuid.New()
	u.Password = string(hash)
	u.CreatedAt = time.Now()

	return s.repo.Create(ctx, u)
}

func (s *userService) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *userService) GetByUsername(ctx context.Context, username string) (*User, error) {
	return s.repo.GetByUsername(ctx, username)
}

func (s *userService) Update(ctx context.Context, id uuid.UUID, dto UserDTO) error {
	if dto.Username != nil {
		exists, err := s.repo.ExistsByUsername(ctx, *dto.Username)
		if err != nil {
			return err
		}
		if exists {
			return errors.New("username taken")
		}
	}

	return s.repo.Update(ctx, id, dto)
}

func (s *userService) UpdatePassword(ctx context.Context, id uuid.UUID, dto UpdatePasswordDTO) error {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if !s.CheckPassword(u, dto.OldPassword) {
		return errors.New("incorrect password")
	}

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(dto.NewPassword),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return err
	}

	return s.repo.UpdatePassword(ctx, id, string(hash))
}

func (s *userService) CheckPassword(u *User, password string) bool {
	return bcrypt.CompareHashAndPassword(
		[]byte(u.Password),
		[]byte(password),
	) == nil
}
