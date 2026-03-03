package user

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo      *UserRepository
	jwtSecret string
}

func NewUserService(repo *UserRepository, jwtSecret string) *UserService {
	return &UserService{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func (s *UserService) Register(ctx context.Context, email, password, name string) (string, error) {
	exists, err := s.repo.EmailExists(ctx, email)
	if err != nil {
		return "", err
	}
	if exists {
		return "", errors.New("email already registered")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	user := &User{
		ID:        uuid.New().String(),
		Email:     email,
		Name:      name,
	}

	if err := s.repo.Create(ctx, user, string(hashedPassword)); err != nil {
		return "", err
	}

	return user.ID, nil
}

func (s *UserService) Login(ctx context.Context, email, password string) (string, *User, error) {
	user, hashedPassword, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return "", nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return "", nil, errors.New("invalid credentials")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(JWTExpiryDuration).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", nil, err
	}

	return tokenString, user, nil
}

func (s *UserService) GetUser(ctx context.Context, userID string) (*User, error) {
	return s.repo.GetByID(ctx, userID)
}
