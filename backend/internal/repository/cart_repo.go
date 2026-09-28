package repository

import (
	"database/sql"

	"ecommerce/backend/internal/models"
)

type CartRepository struct {
	db *sql.DB
}

func NewCartRepository(db *sql.DB) *CartRepository {
	return &CartRepository{db: db}
}

// ensureCart returns the cart id for a user, creating one if needed.
func (r *CartRepository) ensureCart(userID int64) (int64, error) {
	var cartID int64
	err := r.db.QueryRow(`SELECT id FROM carts WHERE user_id = $1`, userID).Scan(&cartID)
	if err == sql.ErrNoRows {
		err = r.db.QueryRow(`INSERT INTO carts (user_id) VALUES ($1) RETURNING id`, userID).Scan(&cartID)
	}
	if err != nil {
		return 0, err
	}
	return cartID, nil
}

func (r *CartRepository) Get(userID int64) (*models.Cart, error) {
	rows, err := r.db.Query(`
		SELECT ci.id, ci.product_id, p.name, p.slug, p.image_url, p.price, ci.quantity, p.stock, ci.created_at
		FROM cart_items ci
		JOIN carts c ON c.id = ci.cart_id
		JOIN products p ON p.id = ci.product_id
		WHERE c.user_id = $1
		ORDER BY ci.created_at ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cart := &models.Cart{Items: []models.CartItem{}}
	for rows.Next() {
		var item models.CartItem
		if err := rows.Scan(&item.ID, &item.ProductID, &item.Name, &item.Slug, &item.ImageURL,
			&item.Price, &item.Quantity, &item.Stock, &item.CreatedAt); err != nil {
			return nil, err
		}
		item.Subtotal = item.Price * float64(item.Quantity)
		cart.Total += item.Subtotal
		cart.Items = append(cart.Items, item)
	}
	return cart, nil
}

func (r *CartRepository) AddItem(userID, productID int64, quantity int) error {
	cartID, err := r.ensureCart(userID)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(`
		INSERT INTO cart_items (cart_id, product_id, quantity)
		VALUES ($1, $2, $3)
		ON CONFLICT (cart_id, product_id)
		DO UPDATE SET quantity = cart_items.quantity + EXCLUDED.quantity, updated_at = now()
	`, cartID, productID, quantity)
	return err
}

func (r *CartRepository) UpdateItem(userID, productID int64, quantity int) error {
	res, err := r.db.Exec(`
		UPDATE cart_items SET quantity = $1, updated_at = now()
		WHERE product_id = $2 AND cart_id = (SELECT id FROM carts WHERE user_id = $3)
	`, quantity, productID, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *CartRepository) RemoveItem(userID, productID int64) error {
	res, err := r.db.Exec(`
		DELETE FROM cart_items
		WHERE product_id = $1 AND cart_id = (SELECT id FROM carts WHERE user_id = $2)
	`, productID, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *CartRepository) Clear(userID int64) error {
	_, err := r.db.Exec(`
		DELETE FROM cart_items WHERE cart_id = (SELECT id FROM carts WHERE user_id = $1)
	`, userID)
	return err
}

// ClearTx clears the cart within an existing transaction (used during
// checkout so the cart is emptied atomically with order creation).
func (r *CartRepository) ClearTx(tx *sql.Tx, userID int64) error {
	_, err := tx.Exec(`
		DELETE FROM cart_items WHERE cart_id = (SELECT id FROM carts WHERE user_id = $1)
	`, userID)
	return err
}
