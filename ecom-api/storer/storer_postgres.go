package storer

import (
	"context"
	"database/sql"
	"fmt"
)

type PostgresStorer struct {
	db *sql.DB
}

func NewPostgresStorer(db *sql.DB) *PostgresStorer {
	return &PostgresStorer{
		db: db,
	}
}

func (ms *PostgresStorer) CreateProduct(ctx context.Context, p *Product) (*Product, error) {
	query := "INSERT INTO products (name, image, category, description, rating, num_reviews, price, count_in_stock) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id"
	err := ms.db.QueryRowContext(ctx, query,
		p.Name, p.Image, p.Category, p.Description,
		p.Rating, p.NumReviews, p.Price, p.CountInStock).Scan(&p.ID)
	if err != nil {
		return nil, fmt.Errorf("error inserting product: %w", err)
	}

	return p, nil
}

func (ms *PostgresStorer) GetProduct(ctx context.Context, id int64) (*Product, error) {
	var p Product
	err := ms.db.QueryRowContext(ctx, "SELECT id, name, image, category, description, rating, num_reviews, price, count_in_stock FROM products WHERE id = $1", id).Scan(
		&p.ID,
		&p.Name,
		&p.Image,
		&p.Category,
		&p.Description,
		&p.Rating,
		&p.NumReviews,
		&p.Price,
		&p.CountInStock,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting product: %w", err)
	}

	return &p, nil
}

func (ms *PostgresStorer) ListProducts(ctx context.Context) ([]Product, error) {
	var products []Product

	rows, err := ms.db.QueryContext(ctx, "SELECT id, name, image, category, description, rating, num_reviews, price, count_in_stock FROM products")
	if err != nil {
		return nil, fmt.Errorf("error querying products: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var p Product
		err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Image,
			&p.Category,
			&p.Description,
			&p.Rating,
			&p.NumReviews,
			&p.Price,
			&p.CountInStock,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning product: %w", err)
		}
		products = append(products, p)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating products: %w", err)
	}

	return products, nil
}

func (ms *PostgresStorer) UpdateProduct(ctx context.Context, p *Product) (*Product, error) {
	_, err := ms.db.ExecContext(ctx,
		"UPDATE products SET name=$1, image=$2, category=$3, description=$4, rating=$5, num_reviews=$6, price=$7, count_in_stock=$8, updated_at=$9 WHERE id=$10",
		p.Name,
		p.Image,
		p.Category,
		p.Description,
		p.Rating,
		p.NumReviews,
		p.Price,
		p.CountInStock,
		p.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("error updating product: %w", err)
	}

	return p, nil
}

func (ms *PostgresStorer) DeleteProduct(ctx context.Context, id int64) error {
	query := "DELETE FROM products WHERE id=$1"
	_, err := ms.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting product: %w", err)
	}

	return nil
}

func (ms *PostgresStorer) CreateOrder(ctx context.Context, o *Order) (*Order, error) {
	err := ms.execTx(ctx, func(tx *sql.Tx) error {
		order, err := createOrder(ctx, tx, o)
		if err != nil {
			return fmt.Errorf("error creating order: %w", err)
		}

		for _, oi := range o.Items {
			oi.OrderID = order.ID
			err = createOrderItem(ctx, tx, oi)
			if err != nil {
				return fmt.Errorf("error creating order item: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error creating order: %w", err)
	}

	return o, nil
}

func createOrder(ctx context.Context, tx *sql.Tx, o *Order) (*Order, error) {
	res, err := tx.ExecContext(ctx, "INSERT INTO orders (payment_method, tax_price, shipping_price, total_price) VALUES (:payment_method, :tax_price, :shipping_price, :total_price)", o)
	if err != nil {
		return nil, fmt.Errorf("error inserting order: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("error getting last insert ID: %w", err)
	}

	o.ID = id

	return o, nil
}

func createOrderItem(ctx context.Context, tx *sql.Tx, oi OrderItem) error {
	res, err := tx.ExecContext(ctx, "INSERT INTO order_items (name, quantity, image, price, product_id, order_id) VALUES (:name, :quantity, :image, :price, :product_id, :order_id)", oi)
	if err != nil {
		return fmt.Errorf("error inserting order item: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert ID: %w", err)
	}

	oi.ID = id

	return nil
}

func (ms *PostgresStorer) GetOrder(ctx context.Context, id int64) (*Order, error) {
	var o Order
	// Use QueryRowContext for fetching a single row
	err := ms.db.QueryRowContext(ctx, "SELECT id, payment_method, tax_price, shipping_price, total_price FROM orders WHERE id = $1", id).Scan(
		&o.ID,
		&o.PaymentMethod,
		&o.TaxPrice,
		&o.ShippingPrice,
		&o.TotalPrice,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found: %w", err)
		}
		return nil, fmt.Errorf("error getting order: %w", err)
	}

	// Fetch associated order items
	rows, err := ms.db.QueryContext(ctx, "SELECT id, order_id, product_id, quantity, price FROM order_items WHERE order_id = $1", id)
	if err != nil {
		return nil, fmt.Errorf("error querying order items: %w", err)
	}
	defer rows.Close()

	var items []OrderItem
	for rows.Next() {
		var item OrderItem
		err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.Quantity,
			&item.Price,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning order item: %w", err)
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating order items: %w", err)
	}

	o.Items = items
	return &o, nil
}

func (ms *PostgresStorer) ListOrders(ctx context.Context) ([]Order, error) {
	var orders []Order

	// Execute the query and retrieve rows
	rows, err := ms.db.QueryContext(ctx, "SELECT id, payment_method, tax_price, shipping_price, total_price FROM orders")
	if err != nil {
		return nil, fmt.Errorf("error listing orders: %w", err)
	}
	defer rows.Close()

	// Iterate over rows and populate the orders slice
	for rows.Next() {
		var order Order
		err := rows.Scan(
			&order.ID,
			&order.PaymentMethod,
			&order.TaxPrice,
			&order.ShippingPrice,
			&order.TotalPrice,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning order: %w", err)
		}
		orders = append(orders, order)
	}

	// Check for errors during iteration
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating orders: %w", err)
	}

	// Fetch order items for each order
	for i := range orders {
		var items []OrderItem
		itemRows, err := ms.db.QueryContext(ctx, "SELECT id, order_id, product_id, quantity, price FROM order_items WHERE order_id = $1", orders[i].ID)
		if err != nil {
			return nil, fmt.Errorf("error getting order items: %w", err)
		}
		defer itemRows.Close()

		for itemRows.Next() {
			var item OrderItem
			err := itemRows.Scan(
				&item.ID,
				&item.OrderID,
				&item.ProductID,
				&item.Quantity,
				&item.Price,
			)
			if err != nil {
				return nil, fmt.Errorf("error scanning order item: %w", err)
			}
			items = append(items, item)
		}

		if err = itemRows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating order items: %w", err)
		}

		orders[i].Items = items
	}

	return orders, nil
}

func (ms *PostgresStorer) DeleteOrder(ctx context.Context, id int64) error {
	// Use a transaction to ensure both the order and its items are deleted atomically
	return ms.execTx(ctx, func(tx *sql.Tx) error {
		// Delete associated order items first
		_, err := tx.ExecContext(ctx, "DELETE FROM order_items WHERE order_id = $1", id)
		if err != nil {
			return fmt.Errorf("error deleting order items: %w", err)
		}

		// Delete the order
		_, err = tx.ExecContext(ctx, "DELETE FROM orders WHERE id = $1", id)
		if err != nil {
			return fmt.Errorf("error deleting order: %w", err)
		}

		return nil
	})
}

func (ms *PostgresStorer) execTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := ms.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}

	err = fn(tx)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("error rolling back transaction: %w", rbErr)
		}
		return fmt.Errorf("error in transaction: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}
