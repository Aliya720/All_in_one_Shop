package handlers

import (
	"net/http"
	"strings"

	"ecommerce/backend/internal/auth"
	"ecommerce/backend/internal/config"
	"ecommerce/backend/internal/middleware"
	"ecommerce/backend/internal/models"
	"ecommerce/backend/internal/repository"
	"ecommerce/backend/internal/utils"
)

type AuthHandler struct {
	users *repository.UserRepository
	cfg   *config.Config
}

func NewAuthHandler(users *repository.UserRepository, cfg *config.Config) *AuthHandler {
	return &AuthHandler{users: users, cfg: cfg}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if req.Name == "" {
		utils.Error(w, http.StatusBadRequest, "name is required")
		return
	}
	if !utils.IsValidEmail(req.Email) {
		utils.Error(w, http.StatusBadRequest, "a valid email is required")
		return
	}
	if !utils.IsValidPassword(req.Password) {
		utils.Error(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "could not process password")
		return
	}

	user, err := h.users.Create(req.Name, req.Email, hash, models.RoleCustomer)
	if err != nil {
		if err == repository.ErrDuplicate {
			utils.Error(w, http.StatusConflict, "an account with this email already exists")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "could not create account")
		return
	}

	token, err := auth.GenerateToken(h.cfg.JWTSecret, h.cfg.JWTExpiryHours, *user)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "could not generate token")
		return
	}

	utils.JSON(w, http.StatusCreated, models.AuthResponse{Token: token, User: *user})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	user, err := h.users.FindByEmail(req.Email)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	if !auth.CheckPassword(user.PasswordHash, req.Password) {
		utils.Error(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	token, err := auth.GenerateToken(h.cfg.JWTSecret, h.cfg.JWTExpiryHours, *user)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "could not generate token")
		return
	}

	utils.JSON(w, http.StatusOK, models.AuthResponse{Token: token, User: *user})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	user, err := h.users.FindByID(userID)
	if err != nil {
		utils.Error(w, http.StatusNotFound, "user not found")
		return
	}

	utils.JSON(w, http.StatusOK, user)
}
