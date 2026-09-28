package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"ecommerce/backend/internal/models"
	"ecommerce/backend/internal/repository"
	"ecommerce/backend/internal/utils"

	"github.com/gorilla/mux"
)

type CategoryHandler struct {
	categories *repository.CategoryRepository
}

func NewCategoryHandler(categories *repository.CategoryRepository) *CategoryHandler {
	return &CategoryHandler{categories: categories}
}

func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	categories, err := h.categories.List()
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "could not list categories")
		return
	}
	utils.JSON(w, http.StatusOK, categories)
}

func (h *CategoryHandler) Get(w http.ResponseWriter, r *http.Request) {
	slug := mux.Vars(r)["slug"]
	category, err := h.categories.FindBySlug(slug)
	if err != nil {
		utils.Error(w, http.StatusNotFound, "category not found")
		return
	}
	utils.JSON(w, http.StatusOK, category)
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in models.CategoryInput
	if err := utils.DecodeJSON(r, &in); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		utils.Error(w, http.StatusBadRequest, "name is required")
		return
	}

	slug := utils.Slugify(in.Name)
	category, err := h.categories.Create(in.Name, slug, in.Description)
	if err != nil {
		if err == repository.ErrDuplicate {
			utils.Error(w, http.StatusConflict, "a category with this name already exists")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "could not create category")
		return
	}

	utils.JSON(w, http.StatusCreated, category)
}

func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid category id")
		return
	}

	var in models.CategoryInput
	if err := utils.DecodeJSON(r, &in); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		utils.Error(w, http.StatusBadRequest, "name is required")
		return
	}

	category, err := h.categories.Update(id, in.Name, in.Description)
	if err != nil {
		if err == repository.ErrNotFound {
			utils.Error(w, http.StatusNotFound, "category not found")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "could not update category")
		return
	}

	utils.JSON(w, http.StatusOK, category)
}

func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid category id")
		return
	}

	if err := h.categories.Delete(id); err != nil {
		switch err {
		case repository.ErrNotFound:
			utils.Error(w, http.StatusNotFound, "category not found")
		case repository.ErrInUse:
			utils.Error(w, http.StatusConflict, "category has products assigned and cannot be deleted")
		default:
			utils.Error(w, http.StatusInternalServerError, "could not delete category")
		}
		return
	}

	utils.JSON(w, http.StatusOK, utils.MessageResponse{Message: "category deleted"})
}
