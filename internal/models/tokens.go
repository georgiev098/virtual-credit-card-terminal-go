package models

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"time"
)

const (
	ScopeAuthentication = "authentication"
)

type Token struct {
	PlainText string    `json:"token"`
	UserId    int64     `json:"-"`
	Hash      []byte    `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

func GenerateToken(userID int, ttl time.Duration, scope string) (*Token, error) {
	token := &Token{
		UserId: int64(userID),
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	randomBytes := make([]byte, 16)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	token.PlainText = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	hash := sha256.Sum256([]byte(token.PlainText))
	token.Hash = hash[:]
	return token, nil
}

func (m *DBModel) InsertToken(token *Token, u User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// delete existing tokens
	stmt := `DELETE FROM tokens WHERE user_id = ?`
	_, err := m.DB.ExecContext(ctx, stmt, u.ID)
	if err != nil {
		return err
	}

	stmt = `INSERT INTO tokens (user_id, name, email, token_hash, created_at, updated_at, expiry_date) 
	values (?, ?, ?, ?, ?, ?, ?)
	`
	_, err = m.DB.ExecContext(ctx, stmt, u.ID, u.LastName, u.Email, token.Hash, time.Now(), time.Now(), token.Expiry)
	if err != nil {
		return err
	}

	return nil
}

func (m *DBModel) GetUserByToken(token string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tokenHash := sha256.Sum256([]byte(token))
	var user User

	stmt := `SELECT u.id, u.first_name, u.last_name, u.email
	 FROM users u 
	 INNER JOIN tokens t ON (u.id = t.user_id)  
	 WHERE t.token_hash = ?
	 AND t.expiry_date > ? 
	 `

	err := m.DB.QueryRowContext(ctx, stmt, tokenHash[:], time.Now()).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil

}
