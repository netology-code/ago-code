package films

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
)

var ErrNotFound = errors.New("not found")

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

type Film struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	Rating      float64 `json:"rating"`
	Description string  `json:"description"`
}

func (s *Service) Top(ctx context.Context) ([]*Film, error) {
	items := make([]*Film, 0)

	rows, err := s.pool.Query(ctx, `SELECT id, title, rating, description FROM films ORDER BY rating DESC LIMIT 50`)
	if err != nil {
		if err == pgx.ErrNoRows {
			return items, nil
		}
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		item := &Film{}
		err = rows.Scan(&item.ID, &item.Title, &item.Rating, &item.Description)
		if err != nil {
			log.Print(err)
			return nil, err
		}
		items = append(items, item)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (s *Service) ByID(ctx context.Context, id int64) (*Film, error) {
	item := &Film{ID: id}
	err := s.pool.QueryRow(
		ctx,
		`SELECT id, title, rating, description FROM films WHERE id = $1`,
		id,
	).Scan(&item.ID, &item.Title, &item.Rating, &item.Description)

	if err != nil {
		return nil, err
	}

	return item, err
}

func (s *Service) SearchByRating(ctx context.Context, rating float64) ([]*Film, error) {
	items := make([]*Film, 0)

	rows, err := s.pool.Query(ctx, `SELECT id, title, rating, description FROM films WHERE rating >= $1 ORDER BY rating DESC LIMIT 50`, rating)
	if err != nil {
		if err == pgx.ErrNoRows {
			return items, nil
		}
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		item := &Film{}
		err = rows.Scan(&item.ID, &item.Title, &item.Rating, &item.Description)
		if err != nil {
			log.Print(err)
			return nil, err
		}
		items = append(items, item)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (s *Service) Save(ctx context.Context, itemToSave *Film) (*Film, error) {
	if itemToSave.ID == 0 {
		err := s.pool.QueryRow(
			ctx,
			`INSERT INTO films (title, rating, description) VALUES($1, $2, $3) RETURNING id`,
			itemToSave.Title, itemToSave.Rating, itemToSave.Description,
		).Scan(&itemToSave.ID)
		if err != nil {
			return nil, err
		}
		return itemToSave, nil
	}

	tag, err := s.pool.Exec(
		ctx,
		`UPDATE films SET title = $2, rating = $4, description = $5 WHERE id = $1`,
		itemToSave.ID, itemToSave.Title, itemToSave.Rating, itemToSave.Description,
	)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() != 1 {
		return nil, ErrNotFound
	}
	return itemToSave, nil
}
