package security

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var ErrUserNotFound = errors.New("user not found")

const (
	RoleAdmin = "ADMIN"
	RoleUser = "USER"
)

type Service struct {
	pool *pgxpool.Pool
}

type UserDetails struct {
	ID    int64
	Login string
	Roles []string
	// TODO: остальные поля
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

// Возвращает профиль пользователя по id
func (s *Service) UserDetails(ctx context.Context, id *string) (interface{}, error) {
	details := &UserDetails{}
	err := s.pool.QueryRow(ctx, `
		SELECT u.id, u.login, u.roles FROM tokens t JOIN users u ON t.userId = u.id WHERE t.id = $1
	`, id).Scan(&details.ID, &details.Login, &details.Roles)
	if err != nil {
		if err != pgx.ErrNoRows {
			return nil, ErrUserNotFound
		}
		// в ДЗ научимся заворачивать ошибки
		return nil, err
	}

	return details, nil
}

// Проверяет, есть ли у пользователя соответствующая роль
func (s *Service) HasAnyRole(ctx context.Context, userDetails interface{}, roles ...string) bool {
	details, ok := userDetails.(*UserDetails)
	if !ok {
		return false
	}

	for _, role := range roles {
		for _, r := range details.Roles {
			if role == r {
				return true
			}
		}
	}

	return false
}

func (s *Service) Register(ctx context.Context, login string, password string) (*string, error) {
	var userID int64
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	err = s.pool.QueryRow(ctx, `
		INSERT INTO users(login, password, roles) VALUES($1, $2, $3) RETURNING id
	`, login, hash, []string{RoleUser}).Scan(&userID)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 256)
	_, err = rand.Read(buf) // crypto/rand
	if err != nil {
		return nil, err
	}

	token := base64.RawURLEncoding.EncodeToString(buf)

	_, err = s.pool.Exec(ctx, `INSERT INTO tokens (id, userId) VALUES ($1, $2)`, token, userID)
	if err != nil {
		// в ДЗ научимся заворачивать ошибки
		return nil, err
	}

	return &token, nil
}

func (s *Service) Login(ctx context.Context, login string, password string) (*string, error) {
	var userID int64
	var hash []byte
	err := s.pool.QueryRow(ctx, `
		SELECT id, password FROM users WHERE login = $1
	`, login).Scan(&userID, &hash)
	if err != nil {
		if err != pgx.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword(hash, []byte(password))
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 256)
	_, err = rand.Read(buf) // crypto/rand
	if err != nil {
		return nil, err
	}

	token := base64.RawURLEncoding.EncodeToString(buf)

	_, err = s.pool.Exec(ctx, `INSERT INTO tokens (id, userId) VALUES ($1, $2)`, token, userID)
	if err != nil {
		// в ДЗ научимся заворачивать ошибки
		return nil, err
	}

	return &token, nil
}
