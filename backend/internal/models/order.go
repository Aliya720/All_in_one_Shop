package models

import "time"

type OrderStatus string

const (
	OrderPending   OrderStatus = "pending"
	OrderPaid      OrderStatus = "paid"
	OrderShipped   OrderStatus = "shipped"
	OrderDelivered OrderStatus = "delivered"
	OrderCancelled OrderStatus = "cancelled"
)

type OrderItem struct {
	ID          int64   `json:"id"`
	ProductID   int64   `json:"product_id"`
	ProductName string  `json:"product_name"`
	ImageURL    string  `json:"image_url"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
	Subtotal    float64 `json:"subtotal"`
}

type Order struct {
	ID              int64       `json:"id"`
	UserID          int64       `json:"user_id"`
	Status          OrderStatus `json:"status"`
	Total           float64     `json:"total"`
	ShippingAddress string      `json:"shipping_address"`
	Items           []OrderItem `json:"items,omitempty"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

type CheckoutRequest struct {
	ShippingAddress string `json:"shipping_address"`
}

type UpdateOrderStatusRequest struct {
	Status OrderStatus `json:"status"`
}
