package models

import "time"

type CartItem struct {
	ID        int64     `json:"id"`
	ProductID int64     `json:"product_id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	ImageURL  string    `json:"image_url"`
	Price     float64   `json:"price"`
	Quantity  int       `json:"quantity"`
	Stock     int       `json:"stock"`
	Subtotal  float64   `json:"subtotal"`
	CreatedAt time.Time `json:"created_at"`
}

type Cart struct {
	Items []CartItem `json:"items"`
	Total float64    `json:"total"`
}

type AddCartItemRequest struct {
	ProductID int64 `json:"product_id"`
	Quantity  int   `json:"quantity"`
}

type UpdateCartItemRequest struct {
	Quantity int `json:"quantity"`
}
