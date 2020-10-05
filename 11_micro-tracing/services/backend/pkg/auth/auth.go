package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go.opencensus.io/plugin/ochttp/propagation/tracecontext"
	"go.opencensus.io/trace"
	"net/http"
)

// можно обращаться по DNS
const target = "auth.service.consul:9999"

type Service struct {
	client   *http.Client
}

func NewService(client *http.Client) *Service {
	return &Service{client: client}
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


	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("http://%s/api/token", target), bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	(&tracecontext.HTTPFormat{}).SpanContextToRequest(trace.FromContext(ctx).SpanContext(), req)

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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("http://%s/api/id", target), bytes.NewReader(data))
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
