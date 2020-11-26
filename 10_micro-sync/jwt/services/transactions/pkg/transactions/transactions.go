package transactions

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"time"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

type Transaction struct {
	ID       int64
	UserID   int64
	Category string
	Amount   int64
	Created  int64
}

func (s *Service) Transactions(ctx context.Context, userID int64) ([]*Transaction, error) {
	transactions := make([]*Transaction, 0)
	rows, err := s.pool.Query(ctx, `
		SELECT id, userid, category, amount, created FROM transactions WHERE userid = $1 ORDER BY id DESC LIMIT 50
	`, userID)
	if err != nil {
		if err != pgx.ErrNoRows {
			return transactions, nil
		}
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		t := &Transaction{}
		var created time.Time
		err = rows.Scan(&t.ID, &t.UserID, &t.Category, &t.Amount, &created)
		if err != nil {
			return nil, err
		}
		t.Created = created.Unix()
		transactions = append(transactions, t)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return transactions, nil
}
