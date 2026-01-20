package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type DBModel struct {
	DB *sql.DB
}

type Models struct {
	DB DBModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		DB: DBModel{
			DB: db,
		},
	}
}

type Widget struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Image          string    `json:"image"`
	IsRecurring    bool      `json:"is_recurring"`
	PlanID         string    `json:"plan_id"`
	InventoryLevel int       `json:"inventory_level"`
	Price          int       `json:"price"`
	CreatedAt      time.Time `json:"-"`
	UpdateddAt     time.Time `json:"-"`
}

type Order struct {
	ID            int       `json:"id"`
	WidgetID      int       `json:"widget_id"`
	TransactionID int       `json:"transaction_id"`
	CustomerID    int       `json:"customer_id"`
	StatusID      int       `json:"status_id"`
	Quantity      int       `json:"quantity"`
	Amount        int       `json:"amount"`
	CreatedAt     time.Time `json:"-"`
	UpdateddAt    time.Time `json:"-"`
}

type Status struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	CreatedAt  time.Time `json:"-"`
	UpdateddAt time.Time `json:"-"`
}

type TransationStatus struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	CreatedAt  time.Time `json:"-"`
	UpdateddAt time.Time `json:"-"`
}

type User struct {
	ID         int       `json:"id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Email      string    `json:"email"`
	Password   string    `json:"password"`
	CreatedAt  time.Time `json:"-"`
	UpdateddAt time.Time `json:"-"`
}

type Customer struct {
	ID         int       `json:"id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Email      string    `json:"email"`
	CreatedAt  time.Time `json:"-"`
	UpdateddAt time.Time `json:"-"`
}

type Transaction struct {
	ID                  int       `json:"id"`
	Amount              int       `json:"amount"`
	TransactionStatusID int       `json:"transaction_status_id"`
	Currency            string    `json:"currency"`
	LastFour            string    `json:"last_four"`
	BankReturnCode      string    `json:"bank_return_code"`
	Name                string    `json:"name"`
	ExpiryMonth         int       `json:"expiry_month"`
	ExpiryYear          int       `json:"expiry_year"`
	PaymentIntent       string    `json:"payment_intent"`
	PaymentMethod       string    `json:"payment_metho"`
	CreatedAt           time.Time `json:"-"`
	UpdateddAt          time.Time `json:"-"`
}

func (m *DBModel) GetWidget(id int) (Widget, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var widget Widget

	row := m.DB.QueryRowContext(ctx, `
    SELECT 
        id,
        name,
        description,
        COALESCE(image, '') AS image,
		is_recurring,
		plan_id,
        inventory_level,
        price,
        created_at,
        updated_at
    FROM widgets
    WHERE id = ?`, id)

	err := row.Scan(
		&widget.ID,
		&widget.Name,
		&widget.Description,
		&widget.Image,
		&widget.IsRecurring,
		&widget.PlanID,
		&widget.InventoryLevel,
		&widget.Price,
		&widget.CreatedAt,
		&widget.UpdateddAt,
	)
	if err != nil {
		return widget, err
	}

	return widget, nil
}

func (m *DBModel) InsertTransaction(txn Transaction) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `INSERT INTO transactions (amount, currency, last_four, bank_return_code, transaction_status_id, created_at, updated_at, expiry_month, expiry_year, payment_intent, payment_method) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := m.DB.ExecContext(ctx, stmt, txn.Amount, txn.Currency, txn.LastFour, txn.BankReturnCode, txn.TransactionStatusID, time.Now(), time.Now(), txn.ExpiryMonth, txn.ExpiryYear, txn.PaymentIntent, txn.PaymentMethod)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (m *DBModel) InsertNewOrder(order Order) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := "INSERT INTO orders (widget_id, transaction_id, status_id, quantity, amount, created_at, updated_at, customer_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"

	result, err := m.DB.ExecContext(ctx, stmt, order.WidgetID, order.TransactionID, order.StatusID, order.Quantity, order.Amount, time.Now(), time.Now(), order.CustomerID)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (m *DBModel) InsertNewCustomer(customer Customer) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := "INSERT INTO customers (first_name, last_name, email, created_at, updated_at) VALUES (?, ?, ?, ?, ?)"

	result, err := m.DB.ExecContext(ctx, stmt, customer.FirstName, customer.LastName, customer.Email, time.Now(), time.Now())
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (m *DBModel) GetUserByEmail(email string) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	emailLower := strings.ToLower(email)
	var u User

	stmt := "SELECT * FROM users WHERE email = ?"

	row := m.DB.QueryRowContext(ctx, stmt, emailLower)

	err := row.Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.Email,
		&u.Password,
		&u.CreatedAt,
		&u.UpdateddAt,
	)
	if err != nil {
		return u, err
	}

	return u, nil
}

func (m *DBModel) Authenticate(email, password string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int
	var hashedPassword string

	row := m.DB.QueryRowContext(ctx, "SELECT id, password FROM users WHERE email = ?", strings.ToLower(email))
	err := row.Scan(&id, &hashedPassword)
	if err != nil {
		return id, err
	}

	fmt.Println("111", hashedPassword)
	// compase passwords
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, errors.New("Incorrect credentials")
	} else if err != nil {
		return 0, err
	}

	return id, nil

}
