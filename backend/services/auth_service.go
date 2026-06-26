package services

import (
	"errors"
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

	// El registro público siempre crea un Cliente. El rol Administrador no se
	// puede auto-asignar desde un campo del request: se provisiona con el seed
	// inicial (ver main.go).
	user := &domain.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: utils.HashPassword(req.Password),
		Role:     domain.RoleClient,
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

func (s *AuthService) ResetPassword(req domain.ResetPasswordRequest) error {
	user, err := s.userDAO.FindByEmail(req.Email)
	if err != nil {
		return errors.New("no existe ningún usuario con ese email")
	}
	return s.userDAO.UpdatePassword(user.ID, utils.HashPassword(req.NewPassword))
}

func (s *AuthService) ChangePassword(userID uint, req domain.ChangePasswordRequest) error {
	user, err := s.userDAO.FindByID(userID)
	if err != nil {
		return errors.New("usuario no encontrado")
	}
	if !utils.CheckPassword(req.CurrentPassword, user.Password) {
		return errors.New("la contraseña actual es incorrecta")
	}
	if len(req.NewPassword) < 6 {
		return errors.New("la nueva contraseña debe tener al menos 6 caracteres")
	}
	return s.userDAO.UpdatePassword(userID, utils.HashPassword(req.NewPassword))
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
