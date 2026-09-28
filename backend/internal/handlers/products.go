package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"ecommerce/backend/internal/config"
	"ecommerce/backend/internal/models"
	"ecommerce/backend/internal/repository"
	"ecommerce/backend/internal/utils"

	"github.com/gorilla/mux"
)

type ProductHandler struct {
	products *repository.ProductRepository
	cfg      *config.Config
}

func NewProductHandler(products *repository.ProductRepository, cfg *config.Config) *ProductHandler {
	return &ProductHandler{products: products, cfg: cfg}
}

func parseFloatPtr(s string) *float64 {
	if s == "" {
		return nil
	}
	if v, err := strconv.ParseFloat(s, 64); err == nil {
		return &v
	}
	return nil
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	filter := repository.ProductFilter{
		Search:       strings.TrimSpace(q.Get("search")),
		CategorySlug: strings.TrimSpace(q.Get("category")),
		MinPrice:     parseFloatPtr(q.Get("min_price")),
		MaxPrice:     parseFloatPtr(q.Get("max_price")),
		Sort:         q.Get("sort"),
		Page:         page,
		Limit:        limit,
		OnlyActive:   true,
	}

	result, err := h.products.List(filter)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "could not list products")
		return
	}
	utils.JSON(w, http.StatusOK, result)
}

// ListAdmin includes inactive products and does not force OnlyActive.
func (h *ProductHandler) ListAdmin(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	filter := repository.ProductFilter{
		Search:       strings.TrimSpace(q.Get("search")),
		CategorySlug: strings.TrimSpace(q.Get("category")),
		Sort:         q.Get("sort"),
		Page:         page,
		Limit:        limit,
		OnlyActive:   false,
	}

	result, err := h.products.List(filter)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "could not list products")
		return
	}
	utils.JSON(w, http.StatusOK, result)
}

func (h *ProductHandler) Get(w http.ResponseWriter, r *http.Request) {
	slug := mux.Vars(r)["slug"]
	product, err := h.products.FindBySlug(slug)
	if err != nil {
		utils.Error(w, http.StatusNotFound, "product not found")
		return
	}
	utils.JSON(w, http.StatusOK, product)
}

func validateProductInput(in models.ProductInput) string {
	if strings.TrimSpace(in.Name) == "" {
		return "name is required"
	}
	if in.Price < 0 {
		return "price cannot be negative"
	}
	if in.Stock < 0 {
		return "stock cannot be negative"
	}
	return ""
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in models.ProductInput
	if err := utils.DecodeJSON(r, &in); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if msg := validateProductInput(in); msg != "" {
		utils.Error(w, http.StatusBadRequest, msg)
		return
	}

	name := strings.TrimSpace(in.Name)
	slug := utils.Slugify(name)

	product, err := h.products.Create(name, slug, in)
	if err != nil {
		if err == repository.ErrDuplicate {
			utils.Error(w, http.StatusConflict, "a product with this name or SKU already exists")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "could not create product")
		return
	}

	utils.JSON(w, http.StatusCreated, product)
}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid product id")
		return
	}

	var in models.ProductInput
	if err := utils.DecodeJSON(r, &in); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if msg := validateProductInput(in); msg != "" {
		utils.Error(w, http.StatusBadRequest, msg)
		return
	}

	name := strings.TrimSpace(in.Name)
	slug := utils.Slugify(name)

	product, err := h.products.Update(id, name, slug, in)
	if err != nil {
		switch err {
		case repository.ErrNotFound:
			utils.Error(w, http.StatusNotFound, "product not found")
		case repository.ErrDuplicate:
			utils.Error(w, http.StatusConflict, "a product with this name or SKU already exists")
		default:
			utils.Error(w, http.StatusInternalServerError, "could not update product")
		}
		return
	}

	utils.JSON(w, http.StatusOK, product)
}

func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid product id")
		return
	}

	if err := h.products.Delete(id); err != nil {
		switch err {
		case repository.ErrNotFound:
			utils.Error(w, http.StatusNotFound, "product not found")
		case repository.ErrInUse:
			utils.Error(w, http.StatusConflict, "product has existing orders and cannot be deleted; consider deactivating it instead")
		default:
			utils.Error(w, http.StatusInternalServerError, "could not delete product")
		}
		return
	}

	utils.JSON(w, http.StatusOK, utils.MessageResponse{Message: "product deleted"})
}

var allowedImageExt = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true,
}

// UploadImage handles multipart/form-data image upload for a product and
// stores the file under the configured upload directory, then updates
// the product's image_url to a path served by the API under /uploads/.
func (h *ProductHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid product id")
		return
	}

	maxBytes := h.cfg.MaxUploadSizeMB * 1024 * 1024
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

	if err := r.ParseMultipartForm(maxBytes); err != nil {
		utils.Error(w, http.StatusBadRequest, fmt.Sprintf("file too large or invalid (max %dMB)", h.cfg.MaxUploadSizeMB))
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "no image file provided (field name must be 'image')")
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedImageExt[ext] {
		utils.Error(w, http.StatusBadRequest, "unsupported image type; allowed: jpg, jpeg, png, webp, gif")
		return
	}

	filename := fmt.Sprintf("product-%d-%d%s", id, time.Now().UnixNano(), ext)
	destPath := filepath.Join(h.cfg.UploadDir, filename)

	dest, err := os.Create(destPath)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "could not save image")
		return
	}
	defer dest.Close()

	if _, err := io.Copy(dest, file); err != nil {
		utils.Error(w, http.StatusInternalServerError, "could not save image")
		return
	}

	imageURL := "/uploads/" + filename
	if err := h.products.UpdateImage(id, imageURL); err != nil {
		if err == repository.ErrNotFound {
			utils.Error(w, http.StatusNotFound, "product not found")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "could not update product image")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"image_url": imageURL})
}
