package repository

import (
	"database/sql"

	"ecommerce/backend/internal/models"
)

type CategoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) Create(name, slug, description string) (*models.Category, error) {
	var c models.Category
	err := r.db.QueryRow(`
		INSERT INTO categories (name, slug, description)
		VALUES ($1, $2, $3)
		RETURNING id, name, slug, description, created_at
	`, name, slug, description).Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.CreatedAt)

	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrDuplicate
		}
		return nil, err
	}
	return &c, nil
}

func (r *CategoryRepository) Update(id int64, name, description string) (*models.Category, error) {
	var c models.Category
	err := r.db.QueryRow(`
		UPDATE categories SET name = $1, description = $2
		WHERE id = $3
		RETURNING id, name, slug, description, created_at
	`, name, description, id).Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CategoryRepository) Delete(id int64) error {
	res, err := r.db.Exec(`DELETE FROM categories WHERE id = $1`, id)
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

func (r *CategoryRepository) FindBySlug(slug string) (*models.Category, error) {
	var c models.Category
	err := r.db.QueryRow(`
		SELECT id, name, slug, description, created_at FROM categories WHERE slug = $1
	`, slug).Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CategoryRepository) List() ([]models.Category, error) {
	rows, err := r.db.Query(`SELECT id, name, slug, description, created_at FROM categories ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.CreatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, nil
}
