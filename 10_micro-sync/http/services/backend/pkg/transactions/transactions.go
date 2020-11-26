package transactions

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
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

// just forwards req/response for simplicity
func (s *Service) Transactions(ctx context.Context, userID int64) ([]byte, error) {
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
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}
