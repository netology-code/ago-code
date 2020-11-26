package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type Service struct {
	client *http.Client
	url    string
}

func NewService(client *http.Client, url string) *Service {
	return &Service{client: client, url: url}
}

func (s *Service) Token(ctx context.Context, login string, password string) (token string, err error) {
	// for simplicity just define locally
	type requestDTO struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	type responseDTO struct {
		Token string `json:"token"`
	}

	data, err := json.Marshal(requestDTO{
		Login:    login,
		Password: password,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/api/token", s.url), bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("status not 200")
	}

	var respDTO responseDTO
	err = json.NewDecoder(resp.Body).Decode(&respDTO)
	if err != nil {
		return "", err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			if err == nil {
				err = cerr
			}
		}
	}()

	return respDTO.Token, nil
}

func (s *Service) Auth(ctx context.Context, token string) (userID int64, err error) {
	// for simplicity just define locally
	type requestDTO struct {
		Token string `json:"token"`
	}

	type responseDTO struct {
		UserID int64 `json:"userId"`
	}

	data, err := json.Marshal(requestDTO{
		Token: token,
	})
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/api/id", s.url), bytes.NewReader(data))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return 0, err
	}

	if resp.StatusCode != http.StatusOK {
		return 0, errors.New("status not 200")
	}

	var respDTO responseDTO
	err = json.NewDecoder(resp.Body).Decode(&respDTO)
	if err != nil {
		return 0, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			if err == nil {
				err = cerr
			}
		}
	}()

	return respDTO.UserID, nil
}
