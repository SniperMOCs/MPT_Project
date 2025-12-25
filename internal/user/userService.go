package user

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/SniperMOCs/MPT_Project/internal/storage"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	repo          UserRepository
	avatarStorage storage.AvatarStorage
}

func NewUserService(repo UserRepository) UserService {
	return &userService{
		repo:          repo,
		avatarStorage: storage.NewAvatarStorage("./uploads", "/uploads"),
	}
}

func (s *userService) Create(ctx context.Context, u *User) error {
	if u.Username == " " || u.Password == " " {
		return errors.New("username and password requiered")
	}
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

func (s *userService) Update(ctx context.Context, id uuid.UUID, username string, fullname string, description string) error {
	if username != "" {
		exists, err := s.repo.ExistsByUsername(ctx, username)
		if err != nil {
			return err
		}
		if exists {
			return errors.New("username taken")
		}
	}

	return s.repo.Update(ctx, id, username, fullname, description)
}

func (s *userService) UpdatePassword(ctx context.Context, id uuid.UUID, oldPassword string, newPassword string) error {
	if oldPassword == "" || newPassword == "" {
		return errors.New("both passwords requiered")
	}
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if !s.CheckPassword(u, oldPassword) {
		return errors.New("incorrect password")
	}

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(newPassword),
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

func (s *userService) UpdateAvatar(ctx context.Context, id uuid.UUID, file io.Reader) error {
	path, err := s.avatarStorage.Save(ctx, id, file)
	if err != nil {
		return err
	}

	return s.repo.UpdateAvatar(ctx, id, path)
}
