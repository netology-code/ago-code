package offers

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

type Offer struct {
	ID      int64  `json:"id"`
	Company string `json:"company"`
	Percent string `json:"percent"`
	Comment string `json:"comment"`
}

func (s *Service) All(ctx context.Context) ([]*Offer, error) {
	items := make([]*Offer, 0)

	rows, err := s.pool.Query(ctx, `SELECT id, company, percent, comment FROM offers`)
	if err != nil {
		if err == pgx.ErrNoRows {
			return items, nil
		}

		log.Print(err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		item := &Offer{}
		err = rows.Scan(&item.ID, &item.Company, &item.Percent, &item.Comment)
		if err != nil {
			log.Print(err)
			return nil, err
		}
		items = append(items, item)
	}
	err = rows.Err()
	if err != nil {
		log.Print(err)
		return nil, err
	}

	return items, nil
}

func (s *Service) ByID(ctx context.Context, id int64) (*Offer, error) {
	item := &Offer{ID: id}
	err := s.pool.QueryRow(
		ctx,
		`SELECT company, percent, comment FROM offers WHERE id = $1`,
		id,
	).Scan(&item.Company, &item.Percent, &item.Comment)

	if err != nil {
		// для no rows уже возвращаем ошибку
		return nil, err
	}

	return item, err
}

func (s *Service) Save(ctx context.Context, itemToSave *Offer) (*Offer, error) {
	if itemToSave.ID == 0 {
		err := s.pool.QueryRow(
			ctx,
			`INSERT INTO offers (company, percent, comment) VALUES($1, $2, $3) RETURNING id`,
			itemToSave.Company, itemToSave.Percent, itemToSave.Comment,
		).Scan(&itemToSave.ID);
		if err != nil {
			return nil, err
		}
		return itemToSave, nil
	}

	tag, err := s.pool.Exec(
		ctx,
		`UPDATE offers SET company = $2, percent = $3, comment = $4 WHERE id = $1`,
		itemToSave.ID, itemToSave.Company, itemToSave.Percent, itemToSave.Comment,
	)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() != 1 {
		return nil, errors.New("No rows updated")
	}
	return itemToSave, nil
}
