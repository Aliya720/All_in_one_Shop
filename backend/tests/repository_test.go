// Package repository_test contains integration tests that exercise the
// repository layer against a real PostgreSQL database. These tests are
// skipped automatically unless TEST_DATABASE_URL is set, so `go test
// ./...` remains safe to run without a database (e.g. in minimal CI
// steps), while `make test` / the CI pipeline that provisions Postgres
// will run them for real coverage of the transactional checkout path.
//
// To run locally:
//
//	createdb ecommerce_test
//	export TEST_DATABASE_URL="postgres://postgres:postgres@localhost:5432/ecommerce_test?sslmode=disable"
//	go test ./tests/...
package repository_test

import (
	"database/sql"
	"os"
	"testing"

	"ecommerce/backend/internal/database"
	"ecommerce/backend/internal/models"
	"ecommerce/backend/internal/repository"

	_ "github.com/lib/pq"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("could not open test database: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Fatalf("could not connect to test database: %v", err)
	}

	// Reset schema for a clean slate on every test run.
	if _, err := db.Exec(`
		DROP TABLE IF EXISTS order_items, orders, cart_items, carts, products, categories, users, schema_migrations CASCADE
	`); err != nil {
		t.Fatalf("could not reset schema: %v", err)
	}

	if err := database.RunMigrations(db, "../migrations"); err != nil {
		t.Fatalf("migrations failed: %v", err)
	}

	t.Cleanup(func() { db.Close() })
	return db
}

func TestUserRepository_CreateAndFind(t *testing.T) {
	db := setupTestDB(t)
	users := repository.NewUserRepository(db)

	user, err := users.Create("Test User", "test@example.com", "hashed", models.RoleCustomer)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if user.ID == 0 {
		t.Error("expected user ID to be set")
	}

	found, err := users.FindByEmail("test@example.com")
	if err != nil {
		t.Fatalf("FindByEmail returned error: %v", err)
	}
	if found.Name != "Test User" {
		t.Errorf("found.Name = %q, want %q", found.Name, "Test User")
	}

	// Duplicate email should be rejected.
	if _, err := users.Create("Another", "test@example.com", "hashed", models.RoleCustomer); err != repository.ErrDuplicate {
		t.Errorf("expected ErrDuplicate for duplicate email, got %v", err)
	}
}

func TestCheckoutFlow_HappyPath(t *testing.T) {
	db := setupTestDB(t)

	users := repository.NewUserRepository(db)
	categories := repository.NewCategoryRepository(db)
	products := repository.NewProductRepository(db)
	carts := repository.NewCartRepository(db)
	orders := repository.NewOrderRepository(db, carts)

	user, err := users.Create("Buyer", "buyer@example.com", "hashed", models.RoleCustomer)
	if err != nil {
		t.Fatalf("could not create user: %v", err)
	}

	category, err := categories.Create("Test Category", "test-category", "")
	if err != nil {
		t.Fatalf("could not create category: %v", err)
	}

	product, err := products.Create("Test Product", "test-product", models.ProductInput{
		Description: "A product for testing",
		Price:       100.00,
		Stock:       5,
		SKU:         "TEST-001",
		CategoryID:  &category.ID,
	})
	if err != nil {
		t.Fatalf("could not create product: %v", err)
	}

	if err := carts.AddItem(user.ID, product.ID, 2); err != nil {
		t.Fatalf("could not add item to cart: %v", err)
	}

	order, err := orders.Checkout(user.ID, "123 Test Street")
	if err != nil {
		t.Fatalf("checkout returned error: %v", err)
	}

	if order.Total != 200.00 {
		t.Errorf("order.Total = %.2f, want 200.00", order.Total)
	}
	if len(order.Items) != 1 {
		t.Fatalf("expected 1 order item, got %d", len(order.Items))
	}
	if order.Items[0].Quantity != 2 {
		t.Errorf("order item quantity = %d, want 2", order.Items[0].Quantity)
	}

	// Stock should have been decremented.
	updatedProduct, err := products.FindByID(product.ID)
	if err != nil {
		t.Fatalf("could not reload product: %v", err)
	}
	if updatedProduct.Stock != 3 {
		t.Errorf("product stock = %d, want 3", updatedProduct.Stock)
	}

	// Cart should now be empty.
	cart, err := carts.Get(user.ID)
	if err != nil {
		t.Fatalf("could not reload cart: %v", err)
	}
	if len(cart.Items) != 0 {
		t.Errorf("expected empty cart after checkout, got %d items", len(cart.Items))
	}
}

func TestCheckoutFlow_InsufficientStock(t *testing.T) {
	db := setupTestDB(t)

	users := repository.NewUserRepository(db)
	products := repository.NewProductRepository(db)
	carts := repository.NewCartRepository(db)
	orders := repository.NewOrderRepository(db, carts)

	user, _ := users.Create("Buyer2", "buyer2@example.com", "hashed", models.RoleCustomer)
	product, _ := products.Create("Scarce Product", "scarce-product", models.ProductInput{
		Price: 50.00,
		Stock: 1,
		SKU:   "SCARCE-001",
	})

	// Add 1 to cart (valid), then manually oversell by dropping stock to 0
	// before checkout to simulate a race with another buyer.
	if err := carts.AddItem(user.ID, product.ID, 1); err != nil {
		t.Fatalf("could not add item to cart: %v", err)
	}
	if _, err := db.Exec(`UPDATE products SET stock = 0 WHERE id = $1`, product.ID); err != nil {
		t.Fatalf("could not force stock to 0: %v", err)
	}

	if _, err := orders.Checkout(user.ID, "Some Address"); err != repository.ErrInsufficientStock {
		t.Errorf("expected ErrInsufficientStock, got %v", err)
	}

	// Cart should be untouched since the transaction rolled back.
	cart, err := carts.Get(user.ID)
	if err != nil {
		t.Fatalf("could not reload cart: %v", err)
	}
	if len(cart.Items) != 1 {
		t.Errorf("expected cart to still have 1 item after failed checkout, got %d", len(cart.Items))
	}
}

func TestCheckoutFlow_EmptyCart(t *testing.T) {
	db := setupTestDB(t)

	users := repository.NewUserRepository(db)
	carts := repository.NewCartRepository(db)
	orders := repository.NewOrderRepository(db, carts)

	user, _ := users.Create("Buyer3", "buyer3@example.com", "hashed", models.RoleCustomer)

	if _, err := orders.Checkout(user.ID, "Some Address"); err != repository.ErrEmptyCart {
		t.Errorf("expected ErrEmptyCart, got %v", err)
	}
}
