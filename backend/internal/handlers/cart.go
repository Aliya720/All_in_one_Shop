package handlers

import (
	"net/http"
	"strconv"

	"ecommerce/backend/internal/middleware"
	"ecommerce/backend/internal/models"
	"ecommerce/backend/internal/repository"
	"ecommerce/backend/internal/utils"

	"github.com/gorilla/mux"
)

type CartHandler struct {
	carts    *repository.CartRepository
	products *repository.ProductRepository
}

func NewCartHandler(carts *repository.CartRepository, products *repository.ProductRepository) *CartHandler {
	return &CartHandler{carts: carts, products: products}
}

func (h *CartHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	cart, err := h.carts.Get(userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "could not load cart")
		return
	}
	utils.JSON(w, http.StatusOK, cart)
}

func (h *CartHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	var req models.AddCartItemRequest
	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Quantity < 1 {
		utils.Error(w, http.StatusBadRequest, "quantity must be at least 1")
		return
	}

	product, err := h.products.FindByID(req.ProductID)
	if err != nil {
		utils.Error(w, http.StatusNotFound, "product not found")
		return
	}
	if !product.IsActive {
		utils.Error(w, http.StatusBadRequest, "product is not available")
		return
	}
	if product.Stock < req.Quantity {
		utils.Error(w, http.StatusBadRequest, "insufficient stock available")
		return
	}

	if err := h.carts.AddItem(userID, req.ProductID, req.Quantity); err != nil {
		utils.Error(w, http.StatusInternalServerError, "could not add item to cart")
		return
	}

	cart, err := h.carts.Get(userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "could not load cart")
		return
	}
	utils.JSON(w, http.StatusOK, cart)
}

func (h *CartHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	productID, err := strconv.ParseInt(mux.Vars(r)["productId"], 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid product id")
		return
	}

	var req models.UpdateCartItemRequest
	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Quantity < 1 {
		utils.Error(w, http.StatusBadRequest, "quantity must be at least 1")
		return
	}

	if err := h.carts.UpdateItem(userID, productID, req.Quantity); err != nil {
		if err == repository.ErrNotFound {
			utils.Error(w, http.StatusNotFound, "item not found in cart")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "could not update cart item")
		return
	}

	cart, err := h.carts.Get(userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "could not load cart")
		return
	}
	utils.JSON(w, http.StatusOK, cart)
}

func (h *CartHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	productID, err := strconv.ParseInt(mux.Vars(r)["productId"], 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid product id")
		return
	}

	if err := h.carts.RemoveItem(userID, productID); err != nil {
		if err == repository.ErrNotFound {
			utils.Error(w, http.StatusNotFound, "item not found in cart")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "could not remove cart item")
		return
	}

	cart, err := h.carts.Get(userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "could not load cart")
		return
	}
	utils.JSON(w, http.StatusOK, cart)
}

func (h *CartHandler) Clear(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	if err := h.carts.Clear(userID); err != nil {
		utils.Error(w, http.StatusInternalServerError, "could not clear cart")
		return
	}
	utils.JSON(w, http.StatusOK, utils.MessageResponse{Message: "cart cleared"})
}
