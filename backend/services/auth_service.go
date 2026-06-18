package services

import (
	"errors"
	"strings"
	"ticketek-backend/dao"
	"ticketek-backend/domain"
	"ticketek-backend/utils"
)

type AuthService struct {
	userDAO *dao.UserDAO
}

func NewAuthService(userDAO *dao.UserDAO) *AuthService {
	return &AuthService{userDAO: userDAO}
}

func (s *AuthService) Register(req domain.RegisterRequest) (*domain.AuthResponse, error) {
	existing, _ := s.userDAO.FindByEmail(req.Email)
	if existing != nil {
		return nil, errors.New("el email ya está registrado")
	}

	role := domain.RoleClient
	if strings.HasSuffix(req.Email, "@tickethub.com") {
		role = domain.RoleAdmin
	}

	user := &domain.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: utils.HashPassword(req.Password),
		Role:     role,
	}

	if err := s.userDAO.Create(user); err != nil {
		return nil, errors.New("error al crear usuario")
	}

	token, err := utils.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, errors.New("error generando token")
	}

	return &domain.AuthResponse{Token: token, User: *user}, nil
}

func (s *AuthService) Login(req domain.LoginRequest) (*domain.AuthResponse, error) {
	user, err := s.userDAO.FindByEmail(req.Email)
	if err != nil {
		return nil, errors.New("credenciales inválidas")
	}

	if !utils.CheckPassword(req.Password, user.Password) {
		return nil, errors.New("credenciales inválidas")
	}

	token, err := utils.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, errors.New("error generando token")
	}

	return &domain.AuthResponse{Token: token, User: *user}, nil
}
