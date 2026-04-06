package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Joseph-Koop/json-project/internal/validator"
)

type Gym struct {
	ID        int       `json:"id"`
	Location  string    `json:"location"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"-"`
	Version   int32     `json:"version"`
}

func ValidateGym(v *validator.Validator, gym *Gym) {

	v.Check(gym.Location != "", "location", "Must be provided.")
	v.Check(len(gym.Location) <= 50, "location", "Must not be more than 50 bytes long.")

	v.Check(gym.Name != "", "name", "Must be provided.")
	v.Check(len(gym.Name) <= 50, "name", "Must not be more than 50 bytes long.")
}

type GymModel struct {
	DB *sql.DB
}

func (c GymModel) Insert(gym *Gym) error {

	query := `
        INSERT INTO gyms (location, name)
        VALUES ($1, $2)
        RETURNING id, created_at, version
        `

	args := []any{gym.Location, gym.Name}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.DB.QueryRowContext(ctx, query, args...).Scan(&gym.ID, &gym.CreatedAt, &gym.Version)

}

func (c GymModel) Get(id int) (*Gym, error) {

	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
        SELECT *
        FROM gyms
        WHERE id = $1
      `

	var gym Gym

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, id).Scan(&gym.ID, &gym.Location, &gym.Name, &gym.CreatedAt, &gym.Version)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &gym, nil
}

func (c GymModel) Update(gym *Gym) error {
	query := `
        UPDATE gyms
        SET location = $1, name = $2, version = version + 1
        WHERE id = $3
        RETURNING version
      `
	args := []any{gym.Location, gym.Name, gym.ID}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.DB.QueryRowContext(ctx, query, args...).Scan(&gym.Version)

}

func (c GymModel) Delete(id int) error {

	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
        DELETE FROM gyms
        WHERE id = $1
      `
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := c.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil

}

func (c GymModel) GetAll(location *string, name *string, filters Filters) ([]*Gym, Metadata, error) {

	query := fmt.Sprintf(`
        SELECT  COUNT(*) OVER(), *
        FROM gyms
        WHERE (to_tsvector('simple', location) @@ 
                plainto_tsquery('simple', $1) OR $1 IS NULL)
            AND (to_tsvector('simple', name) @@ 
                plainto_tsquery('simple', $2) OR $2 IS NULL)
        ORDER BY %s %s, id ASC
        LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := c.DB.QueryContext(ctx, query, location, name, filters.limit(), filters.offset())

	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()
	totalRecords := 0

	gyms := []*Gym{}

	for rows.Next() {
		var gym Gym
		err := rows.Scan(&totalRecords, &gym.ID, &gym.Location, &gym.Name, &gym.CreatedAt, &gym.Version)

		if err != nil {
			return nil, Metadata{}, err
		}

		gyms = append(gyms, &gym)
	}

	err = rows.Err()
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := CalculateMetaData(totalRecords, filters.Page, filters.PageSize)

	return gyms, metadata, nil
}
