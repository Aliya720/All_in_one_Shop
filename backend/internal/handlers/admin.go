package handlers

import (
	"net/http"
	"strconv"

	"ecommerce/backend/internal/repository"
	"ecommerce/backend/internal/utils"
)

type AdminHandler struct {
	users *repository.UserRepository
}

func NewAdminHandler(users *repository.UserRepository) *AdminHandler {
	return &AdminHandler{users: users}
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	users, total, err := h.users.List(limit, offset)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "could not list users")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"items": users,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// Health is a simple liveness/readiness endpoint for Docker/Nginx/monitoring.
func Health(w http.ResponseWriter, r *http.Request) {
	utils.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
