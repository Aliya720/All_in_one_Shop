package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"ecommerce/backend/internal/middleware"
	"ecommerce/backend/internal/models"
	"ecommerce/backend/internal/repository"
	"ecommerce/backend/internal/utils"

	"github.com/gorilla/mux"
)

type OrderHandler struct {
	orders *repository.OrderRepository
}

func NewOrderHandler(orders *repository.OrderRepository) *OrderHandler {
	return &OrderHandler{orders: orders}
}

func (h *OrderHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	var req models.CheckoutRequest
	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if strings.TrimSpace(req.ShippingAddress) == "" {
		utils.Error(w, http.StatusBadRequest, "shipping address is required")
		return
	}

	order, err := h.orders.Checkout(userID, req.ShippingAddress)
	if err != nil {
		switch err {
		case repository.ErrEmptyCart:
			utils.Error(w, http.StatusBadRequest, "your cart is empty")
		case repository.ErrInsufficientStock:
			utils.Error(w, http.StatusConflict, "one or more items no longer have sufficient stock; please review your cart")
		default:
			utils.Error(w, http.StatusInternalServerError, "could not complete checkout")
		}
		return
	}

	utils.JSON(w, http.StatusCreated, order)
}

func (h *OrderHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	orders, total, err := h.orders.ListForUser(userID, page, limit)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "could not list orders")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"items": orders,
		"total": total,
	})
}

func (h *OrderHandler) GetMine(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid order id")
		return
	}

	order, err := h.orders.FindByID(id)
	if err != nil {
		utils.Error(w, http.StatusNotFound, "order not found")
		return
	}
	if order.UserID != userID {
		utils.Error(w, http.StatusForbidden, "you do not have access to this order")
		return
	}

	utils.JSON(w, http.StatusOK, order)
}

// --- Admin ---

func (h *OrderHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	status := r.URL.Query().Get("status")

	orders, total, err := h.orders.ListAll(page, limit, status)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "could not list orders")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"items": orders,
		"total": total,
	})
}

func (h *OrderHandler) GetAny(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid order id")
		return
	}
	order, err := h.orders.FindByID(id)
	if err != nil {
		utils.Error(w, http.StatusNotFound, "order not found")
		return
	}
	utils.JSON(w, http.StatusOK, order)
}

var validStatuses = map[models.OrderStatus]bool{
	models.OrderPending: true, models.OrderPaid: true, models.OrderShipped: true,
	models.OrderDelivered: true, models.OrderCancelled: true,
}

func (h *OrderHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid order id")
		return
	}

	var req models.UpdateOrderStatusRequest
	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if !validStatuses[req.Status] {
		utils.Error(w, http.StatusBadRequest, "invalid status value")
		return
	}

	order, err := h.orders.UpdateStatus(id, req.Status)
	if err != nil {
		if err == repository.ErrNotFound {
			utils.Error(w, http.StatusNotFound, "order not found")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "could not update order")
		return
	}

	utils.JSON(w, http.StatusOK, order)
}

func (h *OrderHandler) Stats(w http.ResponseWriter, r *http.Request) {
	totalOrders, totalRevenue, totalProducts, totalUsers, err := h.orders.Stats()
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "could not load stats")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"total_orders":   totalOrders,
		"total_revenue":  totalRevenue,
		"total_products": totalProducts,
		"total_users":    totalUsers,
	})
}
