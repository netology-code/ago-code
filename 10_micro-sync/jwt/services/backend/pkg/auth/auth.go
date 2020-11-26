package auth

import (
	"backend/pkg/jwt/asymmetric"
	"context"
	"errors"
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
	publicKey     []byte
	tokenLifeTime time.Duration
}

func NewService(publicKey []byte, tokenLifeTime time.Duration) *Service {
	return &Service{publicKey: publicKey, tokenLifeTime: tokenLifeTime}
}

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
