package security

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"auth/pkg/jwt/asymmetric"
	"time"
)

var ErrUserNotFound = errors.New("user not found")
var ErrInvalidToken = errors.New("invalid token")
var ErrExpiredToken = errors.New("expired token")

const (
	RoleAdmin = "ADMIN"
	RoleUser  = "USER"
)

type Service struct {
	pool          *pgxpool.Pool
	privateKey    []byte
	publicKey     []byte
	tokenLifeTime time.Duration
}

func NewService(pool *pgxpool.Pool, privateKey []byte, publicKey []byte, tokenLifeTime time.Duration) *Service {
	return &Service{pool: pool, privateKey: privateKey, publicKey: publicKey, tokenLifeTime: tokenLifeTime}
}

// Возвращает профиль пользователя по id
func (s *Service) UserDetails(ctx context.Context, id *string) (interface{}, error) {
	// Теперь можем не ходить в БД
	token := *id
	ok, err := asymmetric.Verify(token, s.publicKey)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrInvalidToken
	}

	var userDetails UserDetails
	err = asymmetric.Decode(token, &userDetails)
	if err != nil {
		return nil, err
	}

	if !asymmetric.IsNotExpired(userDetails.Expire, time.Now()) {
		return nil, ErrExpiredToken
	}

	return &userDetails, nil
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

	now := time.Now()
	details := &UserDetails{
		UserID: userID,
		Login:  login,
		Roles:  []string{RoleUser},
		Issued: now.Unix(),
		Expire: now.Add(s.tokenLifeTime).Unix(),
	}
	token, err := asymmetric.Encode(details, s.privateKey)
	if err != nil {
		return nil, err
	}

	_, err = s.pool.Exec(ctx, `INSERT INTO tokens (id, userId) VALUES ($1, $2)`, token, userID)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (s *Service) Login(ctx context.Context, login string, password string) (*string, error) {
	var userID int64
	var hash []byte
	var roles []string
	err := s.pool.QueryRow(ctx, `
		SELECT id, password, roles FROM users WHERE login = $1
	`, login).Scan(&userID, &hash, &roles)
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

	now := time.Now()
	details := &UserDetails{
		UserID: userID,
		Login:  login,
		Roles:  roles,
		Issued: now.Unix(),
		Expire: now.Add(s.tokenLifeTime).Unix(),
	}
	token, err := asymmetric.Encode(details, s.privateKey)
	if err != nil {
		return nil, err
	}

	_, err = s.pool.Exec(ctx, `INSERT INTO tokens (id, userId) VALUES ($1, $2)`, token, userID)
	if err != nil {
		// в ДЗ научимся заворачивать ошибки
		return nil, err
	}

	return &token, nil
}
