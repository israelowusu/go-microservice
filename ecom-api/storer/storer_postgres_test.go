package storer

// import (
// 	"context"
// 	"database/sql"
// 	"testing"

// 	"github.com/pashagolub/pgxmock/v4"
// 	"github.com/stretchr/testify/assert"
// )

// func withTestDB(t *testing.T, fn func(*sql.DB, pgxmock.PgxPoolIface)) {
// 	mockDB, mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
// 	if err != nil {
// 		t.Fatalf("error creating mock database: %v", err)
// 	}
// 	defer mockDB.Close()

// 	db := pgxmock.NewPool(mockDB, "pgxmock")
// 	fn(db, mock)
// }

// func TestCreateProduct(t *testing.T) {
// 	// Create mock database
// 	mockDB, err := pgxmock.NewPool()
// 	if err != nil {
// 		t.Fatalf("error creating mock database: %v", err)
// 	}
// 	defer mockDB.Close()

// 	// Create storer with mock db
// 	st := &PostgresStorer{db: mockDB}

// 	// Test product
// 	p := &Product{
// 		Name:         "test product",
// 		Image:        "test.jpg",
// 		Category:     "test category",
// 		Description:  "test description",
// 		Rating:       5,
// 		NumReviews:   10,
// 		Price:        100.0,
// 		CountInStock: 100,
// 	}

// 	// Expected query and parameters
// 	query := "INSERT INTO products (name, image, category, description, rating, num_reviews, price, count_in_stock) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id"

// 	// Set up expectation
// 	mockDB.ExpectQuery(query).
// 		WithArgs(p.Name, p.Image, p.Category, p.Description, p.Rating, p.NumReviews, p.Price, p.CountInStock).
// 		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1))

// 	// Execute the function
// 	result, err := st.CreateProduct(context.Background(), p)

// 	// Assertions
// 	assert.NoError(t, err)
// 	assert.NotNil(t, result)
// 	assert.Equal(t, int64(1), result.ID)
// 	assert.Equal(t, p.Name, result.Name)
// 	assert.Equal(t, p.Price, result.Price)

// 	// Ensure all expectations were met
// 	if err := mockDB.ExpectationsWereMet(); err != nil {
// 		t.Errorf("there were unfulfilled expectations: %s", err)
// 	}
// }
