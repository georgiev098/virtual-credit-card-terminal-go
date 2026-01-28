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
	UpdatedAt      time.Time `json:"-"`
}

type Order struct {
	ID            int         `json:"id"`
	WidgetID      int         `json:"widget_id"`
	TransactionID int         `json:"transaction_id"`
	CustomerID    int         `json:"customer_id"`
	StatusID      int         `json:"status_id"`
	Quantity      int         `json:"quantity"`
	Amount        int         `json:"amount"`
	CreatedAt     time.Time   `json:"-"`
	UpdatedAt     time.Time   `json:"-"`
	Widget        Widget      `json:"widget"`
	Transaction   Transaction `json:"transaction"`
	Customer      Customer    `json:"customer"`
}

type Status struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type TransationStatus struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type User struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type Customer struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
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
	UpdatedAt           time.Time `json:"-"`
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
		&widget.UpdatedAt,
	)
	if err != nil {
		return widget, err
	}

	return widget, nil
}

func (m *DBModel) GetAllSalesPaginated(pageSize int, page int) ([]*Order, int, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	offest := (page - 1) * pageSize

	var orders []*Order

	query := `
		SELECT
			o.id,
			o.widget_id,
			o.transaction_id,
			o.customer_id,
			o.status_id,
			o.quantity,
			o.amount,
			o.created_at,
			o.updated_at,
			w.id,
			w.name,
			t.id,
			t.amount,
			t.currency,
			t.last_four,
			t.expiry_month,
			t.expiry_year,
			t.payment_intent,
			t.bank_return_code,
			c.id,
			c.first_name,
			c.last_name,
			c.email
		FROM orders o
		LEFT JOIN widgets w ON o.widget_id = w.id
		LEFT JOIN transactions t ON o.transaction_id = t.id
		LEFT JOIN customers c ON o.customer_id = c.id
		WHERE w.is_recurring = 0
		ORDER BY o.created_at DESC
		LIMIT ? OFFSET ?
`

	rows, err := m.DB.QueryContext(ctx, query, pageSize, offest)
	if err != nil {
		return nil, 0, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var o Order
		err = rows.Scan(
			&o.ID,
			&o.WidgetID,
			&o.TransactionID,
			&o.CustomerID,
			&o.StatusID,
			&o.Quantity,
			&o.Amount,
			&o.CreatedAt,
			&o.UpdatedAt,
			&o.Widget.ID,
			&o.Widget.Name,
			&o.Transaction.ID,
			&o.Transaction.Amount,
			&o.Transaction.Currency,
			&o.Transaction.LastFour,
			&o.Transaction.ExpiryMonth,
			&o.Transaction.ExpiryYear,
			&o.Transaction.PaymentIntent,
			&o.Transaction.BankReturnCode,
			&o.Customer.ID,
			&o.Customer.FirstName,
			&o.Customer.LastName,
			&o.Customer.Email,
		)
		if err != nil {
			return nil, 0, 0, err
		}

		orders = append(orders, &o)
	}

	query = `
	SELECT count(o.id)
	FROM orders o
	LEFT JOIN widgets w ON (o.widget_id = w.id)
	WHERE w.is_recurring = 0
	`

	var totalRecords int
	coutRow := m.DB.QueryRowContext(ctx, query)
	err = coutRow.Scan(&totalRecords)
	if err != nil {
		return nil, 0, 0, err
	}

	lastPage := totalRecords / pageSize

	return orders, lastPage, totalRecords, nil
}

func (m *DBModel) GetAllSubscriptionsPaginated(pageSize int, page int) ([]*Order, int, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	offest := (page - 1) * pageSize

	var orders []*Order

	query := `
		SELECT
			o.id,
			o.widget_id,
			o.transaction_id,
			o.customer_id,
			o.status_id,
			o.quantity,
			o.amount,
			o.created_at,
			o.updated_at,
			w.id,
			w.name,
			t.id,
			t.amount,
			t.currency,
			t.last_four,
			t.expiry_month,
			t.expiry_year,
			t.payment_intent,
			t.bank_return_code,
			c.id,
			c.first_name,
			c.last_name,
			c.email
		FROM orders o
		LEFT JOIN widgets w ON o.widget_id = w.id
		LEFT JOIN transactions t ON o.transaction_id = t.id
		LEFT JOIN customers c ON o.customer_id = c.id
		WHERE w.is_recurring = 1
		ORDER BY o.created_at DESC
		LIMIT ? OFFSET ?
`

	rows, err := m.DB.QueryContext(ctx, query, pageSize, offest)
	if err != nil {
		return nil, 0, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var o Order
		err = rows.Scan(
			&o.ID,
			&o.WidgetID,
			&o.TransactionID,
			&o.CustomerID,
			&o.StatusID,
			&o.Quantity,
			&o.Amount,
			&o.CreatedAt,
			&o.UpdatedAt,
			&o.Widget.ID,
			&o.Widget.Name,
			&o.Transaction.ID,
			&o.Transaction.Amount,
			&o.Transaction.Currency,
			&o.Transaction.LastFour,
			&o.Transaction.ExpiryMonth,
			&o.Transaction.ExpiryYear,
			&o.Transaction.PaymentIntent,
			&o.Transaction.BankReturnCode,
			&o.Customer.ID,
			&o.Customer.FirstName,
			&o.Customer.LastName,
			&o.Customer.Email,
		)
		if err != nil {
			return nil, 0, 0, err
		}

		orders = append(orders, &o)
	}

	query = `
	SELECT count(o.id)
	FROM orders o
	LEFT JOIN widgets w ON (o.widget_id = w.id)
	WHERE w.is_recurring = 1
	`

	var totalRecords int
	coutRow := m.DB.QueryRowContext(ctx, query)
	err = coutRow.Scan(&totalRecords)
	if err != nil {
		return nil, 0, 0, err
	}

	lastPage := totalRecords / pageSize

	return orders, lastPage, totalRecords, nil
}

func (m *DBModel) GetOneSaleByID(saleId int) (Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var order Order

	query := `
		SELECT
			o.id,
			o.widget_id,
			o.transaction_id,
			o.customer_id,
			o.status_id,
			o.quantity,
			o.amount,
			o.created_at,
			o.updated_at,
			w.id,
			w.name,
			t.id,
			t.amount,
			t.currency,
			t.last_four,
			t.expiry_month,
			t.expiry_year,
			t.payment_intent,
			t.bank_return_code,
			c.id,
			c.first_name,
			c.last_name,
			c.email
		FROM orders o 
		LEFT JOIN widgets w ON o.widget_id = w.id
		LEFT JOIN transactions t ON o.transaction_id = t.id
		LEFT JOIN customers c ON o.customer_id = c.id
		WHERE o.id = ?
`

	row := m.DB.QueryRowContext(ctx, query, saleId)

	err := row.Scan(
		&order.ID,
		&order.WidgetID,
		&order.TransactionID,
		&order.CustomerID,
		&order.StatusID,
		&order.Quantity,
		&order.Amount,
		&order.CreatedAt,
		&order.UpdatedAt,
		&order.Widget.ID,
		&order.Widget.Name,
		&order.Transaction.ID,
		&order.Transaction.Amount,
		&order.Transaction.Currency,
		&order.Transaction.LastFour,
		&order.Transaction.ExpiryMonth,
		&order.Transaction.ExpiryYear,
		&order.Transaction.PaymentIntent,
		&order.Transaction.BankReturnCode,
		&order.Customer.ID,
		&order.Customer.FirstName,
		&order.Customer.LastName,
		&order.Customer.Email,
	)
	if err != nil {
		return order, err
	}

	return order, nil

}

func (m *DBModel) GetAllSubscriptions() ([]*Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var orders []*Order

	query := `
		SELECT
			o.id,
			o.widget_id,
			o.transaction_id,
			o.customer_id,
			o.status_id,
			o.quantity,
			o.amount,
			o.created_at,
			o.updated_at,
			w.id,
			w.name,
			t.id,
			t.amount,
			t.currency,
			t.last_four,
			t.expiry_month,
			t.expiry_year,
			t.payment_intent,
			t.bank_return_code,
			c.id,
			c.first_name,
			c.last_name,
			c.email
		FROM orders o
		LEFT JOIN widgets w ON o.widget_id = w.id
		LEFT JOIN transactions t ON o.transaction_id = t.id
		LEFT JOIN customers c ON o.customer_id = c.id
		WHERE w.is_recurring = 1
		ORDER BY o.created_at DESC
`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var o Order
		err = rows.Scan(
			&o.ID,
			&o.WidgetID,
			&o.TransactionID,
			&o.CustomerID,
			&o.StatusID,
			&o.Quantity,
			&o.Amount,
			&o.CreatedAt,
			&o.UpdatedAt,
			&o.Widget.ID,
			&o.Widget.Name,
			&o.Transaction.ID,
			&o.Transaction.Amount,
			&o.Transaction.Currency,
			&o.Transaction.LastFour,
			&o.Transaction.ExpiryMonth,
			&o.Transaction.ExpiryYear,
			&o.Transaction.PaymentIntent,
			&o.Transaction.BankReturnCode,
			&o.Customer.ID,
			&o.Customer.FirstName,
			&o.Customer.LastName,
			&o.Customer.Email,
		)
		if err != nil {
			return nil, err
		}

		orders = append(orders, &o)
	}

	return orders, nil

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
		&u.UpdatedAt,
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

func (m *DBModel) UpdateUserPassword(u User, hash string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := "UPDATE users SET password = ? WHERE id = ?"

	_, err := m.DB.ExecContext(ctx, stmt, hash, u.ID)
	if err != nil {
		return err
	}

	return nil
}

func (m *DBModel) UpdateOrderStatus(id int, statusId int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `UPDATE orders SET status_id = ? WHERE id = ?`

	_, err := m.DB.ExecContext(ctx, stmt, statusId, id)
	if err != nil {
		return err
	}
	return nil
}

func (m *DBModel) AllUsers() ([]*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var users []*User

	stmt := `
	 SELECT id, first_name, last_name, email, created_at, updated_at
	 FROM users
	 ORDER BY last_name, first_name
	`

	rows, err := m.DB.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var u User

		err := rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.CreatedAt, &u.UpdatedAt)
		if err != nil {
			return nil, err
		}

		users = append(users, &u)
	}

	return users, nil
}

func (m *DBModel) GetOneUserByID(id int) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	user := &User{}

	stmt := `
	 SELECT id, first_name, last_name, email, created_at, updated_at
	 FROM users
	 WHERE id = ?
	`

	row := m.DB.QueryRowContext(ctx, stmt, id)

	err := row.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err // or nil, nil depending on your design
		}
		return nil, err
	}

	return user, nil
}

func (m *DBModel) EditUser(u User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		UPDATE users
		SET first_name = ?, last_name = ?, email = ?, updated_at = ? 
		WHERE id = ?
	`

	_, err := m.DB.ExecContext(ctx, stmt, u.FirstName, u.LastName, u.Email, time.Now(), u.ID)
	if err != nil {
		return err
	}

	return nil
}

func (m *DBModel) InsertUser(u User, hashedPass string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		INSERT INTO users (first_name, last_name, email, password, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := m.DB.ExecContext(ctx, stmt, u.FirstName, u.LastName, u.Email, hashedPass, time.Now(), time.Now())
	if err != nil {
		return err
	}

	return nil
}

func (m *DBModel) DeleteUserByID(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		DELETE FROM users
		WHERE id = ?
	`

	_, err := m.DB.ExecContext(ctx, stmt, id)
	if err != nil {
		return err
	}

	stmt = `DELETE FROM tokens WHERE user_id = ?`

	_, err = m.DB.ExecContext(ctx, stmt, id)
	if err != nil {
		return err
	}

	return nil
}
