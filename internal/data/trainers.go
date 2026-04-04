package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/mail"
	"strconv"
	"time"

	"github.com/Joseph-Koop/json-project/internal/validator"
)

type Trainer struct {
	ID        int     `json:"id"`
	User_id        int     `json:"user_id"`
	Name      string    `json:"name"`
	Address      string    `json:"address"`
	Phone      int    `json:"phone"`
	Email      string    `json:"email"`
	CreatedAt time.Time `json:"-"`
	Version   int32     `json:"version"`
}

func ValidateTrainer(v *validator.Validator, trainer *Trainer) {

	v.Check(len(strconv.Itoa(trainer.User_id)) > 0, "user_id", "Must be an existing user.")

	v.Check(trainer.Name != "", "name", "Must be provided.")
	v.Check(len(trainer.Name) <= 50, "name", "Must not be more than 50 bytes long.")

	v.Check(trainer.Address != "", "address", "Must be provided.")
	v.Check(len(trainer.Address) <= 50, "address", "Must not be more than 100 bytes long.")

	v.Check(len(strconv.Itoa(trainer.Phone)) == 10, "phone", "Must be 10 digits long.")

	v.Check(trainer.Email != "", "email", "Must be provided.")
	v.Check(len(trainer.Email) <= 50, "email", "Must not be more than 50 bytes long.")
	_, err := mail.ParseAddress(trainer.Email)
	v.Check(err == nil, "email", "Must be a valid email address")
}

type TrainerModel struct {
	DB *sql.DB
}

func (c TrainerModel) Insert(trainer *Trainer) error {

	query := `
        INSERT INTO trainers (user_id, name, address, phone, email)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, created_at, version
        `

	args := []any{trainer.User_id, trainer.Name, trainer.Address, trainer.Phone, trainer.Email}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.DB.QueryRowContext(ctx, query, args...).Scan(&trainer.ID, &trainer.CreatedAt, &trainer.Version)

}

func (c TrainerModel) Get(id int64) (*Trainer, error) {

	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
        SELECT *
        FROM trainers
        WHERE id = $1
      `

	var trainer Trainer

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, id).Scan(&trainer.ID, &trainer.User_id, &trainer.Name, &trainer.Address, &trainer.Phone, &trainer.Email, &trainer.CreatedAt, &trainer.Version)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &trainer, nil
}

func (c TrainerModel) Update(trainer *Trainer) error {
	query := `
        UPDATE trainers
        SET user_id = $1, name = $2, address = $3, phone = $4, email = $5, version = version + 1
        WHERE id = $6
        RETURNING version
      `
	args := []any{trainer.User_id, trainer.Name, trainer.Address, trainer.Phone, trainer.Email, trainer.ID}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.DB.QueryRowContext(ctx, query, args...).Scan(&trainer.Version)

}

func (c TrainerModel) Delete(id int64) error {

	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
        DELETE FROM trainers
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

func (c TrainerModel) GetAll(user_id *int, name *string, address *string, phone *int, email *string, filters Filters) ([]*Trainer, Metadata, error) {

	query := fmt.Sprintf(`
        SELECT  COUNT(*) OVER(), *
        FROM trainers
        WHERE (to_tsvector('simple', name) @@ 
                plainto_tsquery('simple', $1) OR $1 IS NULL)
            AND (to_tsvector('simple', address) @@ 
                plainto_tsquery('simple', $2) OR $2 IS NULL)
            AND (phone = $3 OR $3 IS NULL)
            AND (to_tsvector('simple', email) @@ 
                plainto_tsquery('simple', $4) OR $4 IS NULL)
        ORDER BY %s %s, id ASC
        LIMIT $5 OFFSET $6`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := c.DB.QueryContext(ctx, query, name, address, phone, email, filters.limit(), filters.offset())

	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()
	totalRecords := 0

	trainers := []*Trainer{}

	for rows.Next() {
		var trainer Trainer
		err := rows.Scan(&totalRecords, &trainer.ID, &trainer.User_id, &trainer.Name, &trainer.Address, &trainer.Phone, &trainer.Email, &trainer.CreatedAt, &trainer.Version)

		if err != nil {
			return nil, Metadata{}, err
		}

		trainers = append(trainers, &trainer)
	}

	err = rows.Err()
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := CalculateMetaData(totalRecords, filters.Page, filters.PageSize)

	return trainers, metadata, nil
}
