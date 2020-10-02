package transactions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

type Service struct {
	client *http.Client
	url    string
}

func NewService(client *http.Client, url string) *Service {
	return &Service{client: client, url: url}
}

// just returns []byte for simplicity
func (s *Service) Transactions(ctx context.Context, userID int64) ([]byte, error) {
	type Transaction struct {
		ID       string  `json:"id"`
		UserID   int64  `json:"userId"`
		Category string `json:"category"`
		Amount   int64  `json:"amount"`
		Created  int64  `json:"created"`
	}
	type responseDTO struct {
		Transactions  []Transaction    `json:"transactions"`
		CategoryStats map[string]int64 `json:"categoryStats"`
	}

	respDTO := &responseDTO{
		CategoryStats: make(map[string]int64),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/api/transactions", s.url), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-UserId", strconv.FormatInt(userID, 10))

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("status not 200")
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			if err == nil {
				err = cerr
			}
		}
	}()
	err = json.NewDecoder(resp.Body).Decode(&respDTO.Transactions)
	if err != nil {
		return nil, err
	}

	log.Print(respDTO.Transactions)
	for _, t := range respDTO.Transactions {
		respDTO.CategoryStats[t.Category] += t.Amount
	}

	data, err := json.Marshal(respDTO)
	if err != nil {
		return nil, err
	}

	return data, nil
}
