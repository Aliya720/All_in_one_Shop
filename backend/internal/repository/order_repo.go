package repository

import (
	"database/sql"
	"errors"
	"strconv"

	"ecommerce/backend/internal/models"
)

var ErrEmptyCart = errors.New("cart is empty")

type OrderRepository struct {
	db    *sql.DB
	carts *CartRepository
}

func NewOrderRepository(db *sql.DB, carts *CartRepository) *OrderRepository {
	return &OrderRepository{db: db, carts: carts}
}

// Checkout converts the user's current cart into an order inside a single
// database transaction: it validates stock, decrements it, creates the
// order + order_items, and clears the cart. If anything fails, the whole
// operation is rolled back.
func (r *OrderRepository) Checkout(userID int64, shippingAddress string) (*models.Order, error) {
	cart, err := r.carts.Get(userID)
	if err != nil {
		return nil, err
	}
	if len(cart.Items) == 0 {
		return nil, ErrEmptyCart
	}

	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var total float64
	for _, item := range cart.Items {
		total += item.Subtotal
	}

	var orderID int64
	err = tx.QueryRow(`
		INSERT INTO orders (user_id, status, total, shipping_address)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, userID, models.OrderPending, total, shippingAddress).Scan(&orderID)
	if err != nil {
		return nil, err
	}

	for _, item := range cart.Items {
		if err := DecrementStock(tx, item.ProductID, item.Quantity); err != nil {
			return nil, err
		}

		_, err = tx.Exec(`
			INSERT INTO order_items (order_id, product_id, product_name, image_url, price, quantity)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, orderID, item.ProductID, item.Name, item.ImageURL, item.Price, item.Quantity)
		if err != nil {
			return nil, err
		}
	}

	if err := r.carts.ClearTx(tx, userID); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.FindByID(orderID)
}

func (r *OrderRepository) FindByID(id int64) (*models.Order, error) {
	var o models.Order
	err := r.db.QueryRow(`
		SELECT id, user_id, status, total, shipping_address, created_at, updated_at
		FROM orders WHERE id = $1
	`, id).Scan(&o.ID, &o.UserID, &o.Status, &o.Total, &o.ShippingAddress, &o.CreatedAt, &o.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	items, err := r.itemsFor(id)
	if err != nil {
		return nil, err
	}
	o.Items = items
	return &o, nil
}

func (r *OrderRepository) itemsFor(orderID int64) ([]models.OrderItem, error) {
	rows, err := r.db.Query(`
		SELECT id, product_id, product_name, image_url, price, quantity
		FROM order_items WHERE order_id = $1
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []models.OrderItem{}
	for rows.Next() {
		var it models.OrderItem
		if err := rows.Scan(&it.ID, &it.ProductID, &it.ProductName, &it.ImageURL, &it.Price, &it.Quantity); err != nil {
			return nil, err
		}
		it.Subtotal = it.Price * float64(it.Quantity)
		items = append(items, it)
	}
	return items, nil
}

func (r *OrderRepository) ListForUser(userID int64, page, limit int) ([]models.Order, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var total int
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM orders WHERE user_id = $1`, userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(`
		SELECT id, user_id, status, total, shipping_address, created_at, updated_at
		FROM orders WHERE user_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	orders := []models.Order{}
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Status, &o.Total, &o.ShippingAddress, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, 0, err
		}
		orders = append(orders, o)
	}
	return orders, total, nil
}

func (r *OrderRepository) ListAll(page, limit int, status string) ([]models.Order, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	where := "1=1"
	args := []interface{}{}
	if status != "" {
		where = "status = $1"
		args = append(args, status)
	}

	var total int
	countQ := "SELECT COUNT(*) FROM orders WHERE " + where
	if err := r.db.QueryRow(countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	listQ := "SELECT id, user_id, status, total, shipping_address, created_at, updated_at FROM orders WHERE " + where +
		" ORDER BY created_at DESC LIMIT $" + strconv.Itoa(len(args)-1) + " OFFSET $" + strconv.Itoa(len(args))

	rows, err := r.db.Query(listQ, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	orders := []models.Order{}
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Status, &o.Total, &o.ShippingAddress, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, 0, err
		}
		orders = append(orders, o)
	}
	return orders, total, nil
}

func (r *OrderRepository) UpdateStatus(id int64, status models.OrderStatus) (*models.Order, error) {
	res, err := r.db.Exec(`UPDATE orders SET status = $1, updated_at = now() WHERE id = $2`, status, id)
	if err != nil {
		return nil, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return nil, ErrNotFound
	}
	return r.FindByID(id)
}

func (r *OrderRepository) Stats() (totalOrders int, totalRevenue float64, totalProducts int, totalUsers int, err error) {
	err = r.db.QueryRow(`SELECT COUNT(*) FROM orders`).Scan(&totalOrders)
	if err != nil {
		return
	}
	err = r.db.QueryRow(`SELECT COALESCE(SUM(total), 0) FROM orders WHERE status != 'cancelled'`).Scan(&totalRevenue)
	if err != nil {
		return
	}
	err = r.db.QueryRow(`SELECT COUNT(*) FROM products`).Scan(&totalProducts)
	if err != nil {
		return
	}
	err = r.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&totalUsers)
	return
}
