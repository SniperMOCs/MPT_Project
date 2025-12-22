package auth

import (
	"context"
	"errors"

	"github.com/SniperMOCs/MPT_Project/internal/config"
	"github.com/SniperMOCs/MPT_Project/internal/user"
	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	userService user.UserService
	jwt         *JWTManager
	cfg         *config.Config
}

func NewAuthService(userService user.UserService, jwt *JWTManager, cfg *config.Config) AuthService {
	return &authService{
		userService: userService,
		jwt:         jwt,
		cfg:         cfg,
	}
}

func (s *authService) Login(ctx context.Context, username string, password string) (string, error) {
	u, err := s.userService.GetByUsername(ctx, username)
	if err != nil {
		return "", errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword(
		[]byte(u.Password),
		[]byte(password),
	); err != nil {
		return "", errors.New("invalid credentials")
	}

	return s.jwt.GenerateToken(u.ID.String(), s.cfg.TokenTTL)
}
