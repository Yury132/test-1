package service

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"

	"github.com/Yury132/Golang-Task-1/internal/models"
	"github.com/rs/zerolog"
)

const oauthStateString = "pseudo-random"

type Service interface {
	GetUserInfo(state string, code string) ([]byte, error)
	GetUsersList(ctx context.Context) ([]models.User, error)
	// Проверка на существование пользователя
	CheckUser(ctx context.Context, email string) (bool, error)
	// Создание нового пользователя
	CreateUser(ctx context.Context, name string, email string) error
}

type GoogleAPI interface {
	GetUserInfo(token *oauth2.Token) ([]byte, error)
}

type Storage interface {
	GetUsers(ctx context.Context) ([]models.User, error)
	// Проверка на существование пользователя
	CheckUser(ctx context.Context, email string) (bool, error)
	// Создание нового пользователя
	CreateUser(ctx context.Context, name string, email string) error
}

type service struct {
	logger      zerolog.Logger
	oauthConfig *oauth2.Config
	googleAPI   GoogleAPI
	storage     Storage
}

// Получаем данные о пользователи из Google
func (s *service) GetUserInfo(state string, code string) ([]byte, error) {
	if state != oauthStateString {
		return nil, fmt.Errorf("invalid oauth state")
	}

	token, err := s.oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %s", err.Error())
	}

	contents, err := s.googleAPI.GetUserInfo(token)
	if err != nil {
		return nil, err
	}

	return contents, nil
}

func (s *service) GetUsersList(ctx context.Context) ([]models.User, error) {
	users, err := s.storage.GetUsers(ctx)
	if err != nil {
		return nil, err
	}

	return users, nil
}

// Проверка на существование пользователя
func (s *service) CheckUser(ctx context.Context, email string) (bool, error) {
	check, err := s.storage.CheckUser(ctx, email)
	if err != nil {
		return false, err
	}

	return check, nil
}

// Создание нового пользователя
func (s *service) CreateUser(ctx context.Context, name string, email string) error {
	err := s.storage.CreateUser(ctx, name, email)
	if err != nil {
		return err
	}

	return nil
}

func New(logger zerolog.Logger, oauthConfig *oauth2.Config, googleAPI GoogleAPI, storage Storage) Service {
	return &service{
		logger:      logger,
		oauthConfig: oauthConfig,
		googleAPI:   googleAPI,
		storage:     storage,
	}
}
