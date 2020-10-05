package transactions

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

// можно обращаться по DNS
const target = "transactions.service.consul:9999"

type Service struct {
	client   *http.Client
}

func NewService(client *http.Client) *Service {
	return &Service{client: client}
}

// just forwards req/response for simplicity
func (s *Service) Transactions(ctx context.Context, userID int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://%s/api/transactions", target), nil)
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
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}
