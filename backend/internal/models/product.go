package models

import "time"

type Category struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type Product struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	Description  string    `json:"description"`
	Price        float64   `json:"price"`
	Stock        int       `json:"stock"`
	SKU          string    `json:"sku"`
	CategoryID   *int64    `json:"category_id"`
	CategoryName string    `json:"category_name,omitempty"`
	ImageURL     string    `json:"image_url"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ProductListResult struct {
	Items      []Product `json:"items"`
	Total      int       `json:"total"`
	Page       int       `json:"page"`
	Limit      int       `json:"limit"`
	TotalPages int       `json:"total_pages"`
}

type ProductInput struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	SKU         string  `json:"sku"`
	CategoryID  *int64  `json:"category_id"`
	ImageURL    string  `json:"image_url"`
	IsActive    *bool   `json:"is_active"`
}

type CategoryInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
