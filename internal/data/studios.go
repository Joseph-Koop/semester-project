package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Joseph-Koop/json-project/internal/validator"
)

type Studio struct {
	ID        int       `json:"id"`
	Gym_id    int       `json:"gym_id"`
	Name      string    `json:"name"`
	Access    string    `json:"access"`
	CreatedAt time.Time `json:"-"`
	Version   int32     `json:"version"`
}

func ValidateStudio(v *validator.Validator, studio *Studio) {

	v.Check(len(strconv.Itoa(studio.Gym_id)) > 0, "gym_id", "Must be an existing gym.")

	v.Check(studio.Name != "", "name", "Must be provided.")
	v.Check(len(studio.Name) <= 50, "name", "Must not be more than 50 bytes long.")

	v.Check(studio.Access == "general" || studio.Access == "classes", "access", "Must be one of the valid options.")
}

type StudioModel struct {
	DB *sql.DB
}

func (c StudioModel) Insert(studio *Studio) error {

	query := `
        INSERT INTO studios (gym_id, name, access)
        VALUES ($1, $2, $3)
        RETURNING id, created_at, version
        `

	args := []any{studio.Gym_id, studio.Name, studio.Access}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.DB.QueryRowContext(ctx, query, args...).Scan(&studio.ID, &studio.CreatedAt, &studio.Version)

}

func (c StudioModel) Get(id int) (*Studio, error) {

	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
        SELECT *
        FROM studios
        WHERE id = $1
      `

	var studio Studio

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, id).Scan(&studio.ID, &studio.Gym_id, &studio.Name, &studio.Access, &studio.CreatedAt, &studio.Version)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &studio, nil
}

func (c StudioModel) Update(studio *Studio) error {
	query := `
        UPDATE studios
        SET gym_id = $1, name = $2, access = $3, version = version + 1
        WHERE id = $4
        RETURNING version
      `
	args := []any{studio.Gym_id, studio.Name, studio.Access, studio.ID}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.DB.QueryRowContext(ctx, query, args...).Scan(&studio.Version)

}

func (c StudioModel) Delete(id int) error {

	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
        DELETE FROM studios
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

func (c StudioModel) GetAll(gym_id *int, name *string, access *string, filters Filters) ([]*Studio, Metadata, error) {

	query := fmt.Sprintf(`
        SELECT  COUNT(*) OVER(), *
        FROM studios
        WHERE (gym_id = $1 OR $1 IS NULL)
		AND (to_tsvector('simple', name) @@ 
			plainto_tsquery('simple', $2) OR $2 IS NULL)
		AND (access = $3 OR $3 IS NULL)
        ORDER BY %s %s, id ASC
        LIMIT $4 OFFSET $5`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := c.DB.QueryContext(ctx, query, gym_id, name, access, filters.limit(), filters.offset())

	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()
	totalRecords := 0

	studios := []*Studio{}

	for rows.Next() {
		var studio Studio
		err := rows.Scan(&totalRecords, &studio.ID, &studio.Gym_id, &studio.Name, &studio.Access, &studio.CreatedAt, &studio.Version)

		if err != nil {
			return nil, Metadata{}, err
		}

		studios = append(studios, &studio)
	}

	err = rows.Err()
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := CalculateMetaData(totalRecords, filters.Page, filters.PageSize)

	return studios, metadata, nil
}
