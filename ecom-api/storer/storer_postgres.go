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

func (ms *PostgresStorer) ListProducts(ctx context.Context) ([]*Product, error) {
	var products []*Product

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
		products = append(products, &p)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating products: %w", err)
	}

	return products, nil
}

func (ms *PostgresStorer) UpdateProduct(ctx context.Context, p *Product) (*Product, error) {
	_, err := ms.db.ExecContext(ctx,
		"UPDATE products SET name=$1, image=$2, category=$3, description=$4, rating=$5, num_reviews=$6, price=$7, count_in_stock=$8 WHERE id=$9",
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
