package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"ecommerce/backend/internal/models"
)

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// ProductFilter holds all supported query parameters for product listing.
type ProductFilter struct {
	Search       string
	CategorySlug string
	MinPrice     *float64
	MaxPrice     *float64
	Sort         string // "price_asc", "price_desc", "newest", "name"
	Page         int
	Limit        int
	OnlyActive   bool
}

func (r *ProductRepository) List(f ProductFilter) (*models.ProductListResult, error) {
	where := []string{"1=1"}
	args := []interface{}{}
	argN := 1

	if f.OnlyActive {
		where = append(where, "p.is_active = true")
	}
	if f.Search != "" {
		where = append(where, fmt.Sprintf("(p.name ILIKE $%d OR p.description ILIKE $%d)", argN, argN))
		args = append(args, "%"+f.Search+"%")
		argN++
	}
	if f.CategorySlug != "" {
		where = append(where, fmt.Sprintf("c.slug = $%d", argN))
		args = append(args, f.CategorySlug)
		argN++
	}
	if f.MinPrice != nil {
		where = append(where, fmt.Sprintf("p.price >= $%d", argN))
		args = append(args, *f.MinPrice)
		argN++
	}
	if f.MaxPrice != nil {
		where = append(where, fmt.Sprintf("p.price <= $%d", argN))
		args = append(args, *f.MaxPrice)
		argN++
	}

	whereClause := strings.Join(where, " AND ")

	orderClause := "p.created_at DESC"
	switch f.Sort {
	case "price_asc":
		orderClause = "p.price ASC"
	case "price_desc":
		orderClause = "p.price DESC"
	case "name":
		orderClause = "p.name ASC"
	case "newest":
		orderClause = "p.created_at DESC"
	}

	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 20
	}
	offset := (f.Page - 1) * f.Limit

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		WHERE %s`, whereClause)

	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	listArgs := append(append([]interface{}{}, args...), f.Limit, offset)
	listQuery := fmt.Sprintf(`
		SELECT p.id, p.name, p.slug, p.description, p.price, p.stock, p.sku,
		       p.category_id, COALESCE(c.name, ''), p.image_url, p.is_active,
		       p.created_at, p.updated_at
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d`, whereClause, orderClause, argN, argN+1)

	rows, err := r.db.Query(listQuery, listArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []models.Product{}
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.Price, &p.Stock, &p.SKU,
			&p.CategoryID, &p.CategoryName, &p.ImageURL, &p.IsActive, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, p)
	}

	totalPages := (total + f.Limit - 1) / f.Limit
	if totalPages < 1 {
		totalPages = 1
	}

	return &models.ProductListResult{
		Items:      items,
		Total:      total,
		Page:       f.Page,
		Limit:      f.Limit,
		TotalPages: totalPages,
	}, nil
}

func (r *ProductRepository) FindBySlug(slug string) (*models.Product, error) {
	var p models.Product
	err := r.db.QueryRow(`
		SELECT p.id, p.name, p.slug, p.description, p.price, p.stock, p.sku,
		       p.category_id, COALESCE(c.name, ''), p.image_url, p.is_active,
		       p.created_at, p.updated_at
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		WHERE p.slug = $1
	`, slug).Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.Price, &p.Stock, &p.SKU,
		&p.CategoryID, &p.CategoryName, &p.ImageURL, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepository) FindByID(id int64) (*models.Product, error) {
	var p models.Product
	err := r.db.QueryRow(`
		SELECT p.id, p.name, p.slug, p.description, p.price, p.stock, p.sku,
		       p.category_id, COALESCE(c.name, ''), p.image_url, p.is_active,
		       p.created_at, p.updated_at
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		WHERE p.id = $1
	`, id).Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.Price, &p.Stock, &p.SKU,
		&p.CategoryID, &p.CategoryName, &p.ImageURL, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepository) Create(name, slug string, in models.ProductInput) (*models.Product, error) {
	isActive := true
	if in.IsActive != nil {
		isActive = *in.IsActive
	}

	var p models.Product
	err := r.db.QueryRow(`
		INSERT INTO products (name, slug, description, price, stock, sku, category_id, image_url, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, name, slug, description, price, stock, sku, category_id, image_url, is_active, created_at, updated_at
	`, name, slug, in.Description, in.Price, in.Stock, in.SKU, in.CategoryID, in.ImageURL, isActive,
	).Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.Price, &p.Stock, &p.SKU, &p.CategoryID, &p.ImageURL, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrDuplicate
		}
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepository) Update(id int64, name, slug string, in models.ProductInput) (*models.Product, error) {
	isActive := true
	if in.IsActive != nil {
		isActive = *in.IsActive
	}

	var p models.Product
	err := r.db.QueryRow(`
		UPDATE products SET
			name = $1, slug = $2, description = $3, price = $4, stock = $5,
			sku = $6, category_id = $7, image_url = $8, is_active = $9, updated_at = now()
		WHERE id = $10
		RETURNING id, name, slug, description, price, stock, sku, category_id, image_url, is_active, created_at, updated_at
	`, name, slug, in.Description, in.Price, in.Stock, in.SKU, in.CategoryID, in.ImageURL, isActive, id,
	).Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.Price, &p.Stock, &p.SKU, &p.CategoryID, &p.ImageURL, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrDuplicate
		}
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepository) UpdateImage(id int64, imageURL string) error {
	res, err := r.db.Exec(`UPDATE products SET image_url = $1, updated_at = now() WHERE id = $2`, imageURL, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ProductRepository) Delete(id int64) error {
	res, err := r.db.Exec(`DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		if isForeignKeyViolation(err) {
			return ErrInUse
		}
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// DecrementStock reduces stock for a product, failing if insufficient
// stock is available. Intended to run inside a transaction during
// checkout.
func DecrementStock(tx *sql.Tx, productID int64, quantity int) error {
	res, err := tx.Exec(`
		UPDATE products SET stock = stock - $1, updated_at = now()
		WHERE id = $2 AND stock >= $1
	`, quantity, productID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrInsufficientStock
	}
	return nil
}

var ErrInsufficientStock = fmt.Errorf("insufficient stock")
