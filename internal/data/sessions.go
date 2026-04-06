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

type Session struct {
	ID        int       `json:"id"`
	Class_id  int       `json:"class_id"`
	CreatedAt time.Time `json:"-"`
	Version   int32     `json:"version"`
}

func ValidateSession(v *validator.Validator, session *Session) {

	v.Check(len(strconv.Itoa(session.Class_id)) > 0, "class_id", "Must be an existing class.")
}

type SessionModel struct {
	DB *sql.DB
}

func (c SessionModel) Insert(session *Session) error {

	query := `
        INSERT INTO sessions (class_id)
        VALUES ($1)
        RETURNING id, created_at, version
        `

	args := []any{session.Class_id}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.DB.QueryRowContext(ctx, query, args...).Scan(&session.ID, &session.CreatedAt, &session.Version)

}

func (c SessionModel) Get(id int) (*Session, error) {

	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
        SELECT *
        FROM sessions
        WHERE id = $1
      `

	var session Session

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, id).Scan(&session.ID, &session.Class_id, &session.CreatedAt, &session.Version)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &session, nil
}

func (c SessionModel) Update(session *Session) error {
	query := `
        UPDATE sessions
        SET class_id = $1, version = version + 1
        WHERE id = $2
        RETURNING version
      `
	args := []any{session.Class_id, session.ID}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.DB.QueryRowContext(ctx, query, args...).Scan(&session.Version)

}

func (c SessionModel) Delete(id int) error {

	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
        DELETE FROM sessions
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

func (c SessionModel) GetAll(class_id *int, filters Filters) ([]*Session, Metadata, error) {

	query := fmt.Sprintf(`
        SELECT  COUNT(*) OVER(), *
        FROM sessions
        WHERE (class_id = $1 OR $1 IS NULL)
        ORDER BY %s %s, id ASC
        LIMIT $2 OFFSET $3`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := c.DB.QueryContext(ctx, query, class_id, filters.limit(), filters.offset())

	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()
	totalRecords := 0

	sessions := []*Session{}

	for rows.Next() {
		var session Session
		err := rows.Scan(&totalRecords, &session.ID, &session.Class_id, &session.CreatedAt, &session.Version)

		if err != nil {
			return nil, Metadata{}, err
		}

		sessions = append(sessions, &session)
	}

	err = rows.Err()
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := CalculateMetaData(totalRecords, filters.Page, filters.PageSize)

	return sessions, metadata, nil
}
