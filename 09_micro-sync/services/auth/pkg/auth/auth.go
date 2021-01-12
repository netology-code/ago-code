package auth

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var ErrUserNotFound = errors.New("user not found")
var ErrInvalidPass = errors.New("invalid password")

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

func (s *Service) Login(ctx context.Context, login string, password string) (string, error) {
	var userID int64
	var hash []byte
	err := s.pool.QueryRow(ctx, `
		SELECT id, password FROM users WHERE login = $1
	`, login).Scan(&userID, &hash)
	if err != nil {
		if err != pgx.ErrNoRows {
			return "", ErrUserNotFound
		}
		return "", err
	}

	err = bcrypt.CompareHashAndPassword(hash, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return "", ErrInvalidPass
		}
		return "", err
	}

	token := uuid.New().String()
	_, err = s.pool.Exec(ctx, `INSERT INTO tokens (token, userid) VALUES ($1, $2)`, token, userID)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *Service) UserID(ctx context.Context, token string) (userID int64, err error) {
	err = s.pool.QueryRow(ctx, `
		SELECT userid FROM tokens WHERE token = $1
	`, token).Scan(&userID)
	if err != nil {
		if err != pgx.ErrNoRows {
			return 0, ErrUserNotFound
		}
		return 0, err
	}

	return userID, nil
}

