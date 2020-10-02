package payments

import (
	"context"
	"errors"
	"fmt"
	"io"
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
func (s *Service) Pay(ctx context.Context, userID int64, reader io.Reader) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/api/payments", s.url), reader)
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
