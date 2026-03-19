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

type SessionTime struct {
	ID        int     `json:"id"`
	Class_id      int    `json:"class_id"`
	Day      string    `json:"day"`
	Time      time.Time    `json:"time"`
	Duration      int    `json:"duration"`
	CreatedAt time.Time `json:"-"`
	Version   int32     `json:"version"`
}

func ValidateSessionTime(v *validator.Validator, sessionTime *SessionTime) {

	v.Check(len(strconv.Itoa(sessionTime.Class_id)) > 0, "class_id", "Must be an existing class.")
	
    v.Check(sessionTime.Day == "sun" || sessionTime.Day == "mon" || sessionTime.Day == "tue" || sessionTime.Day == "wed" || sessionTime.Day == "thr" || sessionTime.Day == "fri" || sessionTime.Day == "sat", "day", "Must be one of the valid options.")

	v.Check(!sessionTime.Time.IsZero(), "time", "Must be provided.")
	v.Check(sessionTime.Time.Second() == 0, "time", "Seconds are not allowed.")

	v.Check(len(strconv.Itoa(sessionTime.Duration)) > 0 && len(strconv.Itoa(sessionTime.Duration)) <= 240, "class_id", "Must be between 1 minute and 4 hours.")
}

type SessionTimeModel struct {
	DB *sql.DB
}

func (c SessionTimeModel) Insert(sessionTime *SessionTime) error {

	query := `
        INSERT INTO session_times (class_id, day, time, duration)
        VALUES ($1, $2, $3, $4)
        RETURNING id, created_at, version
        `

	args := []any{sessionTime.Class_id, sessionTime.Day, sessionTime.Time, sessionTime.Duration}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.DB.QueryRowContext(ctx, query, args...).Scan(&sessionTime.ID, &sessionTime.CreatedAt, &sessionTime.Version)

}

func (c SessionTimeModel) Get(id int64) (*SessionTime, error) {

	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
        SELECT *
        FROM session_times
        WHERE id = $1
      `

	var sessionTime SessionTime

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, id).Scan(&sessionTime.ID, &sessionTime.Class_id, &sessionTime.Day, &sessionTime.Time, &sessionTime.Duration, &sessionTime.CreatedAt, &sessionTime.Version)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &sessionTime, nil
}

func (c SessionTimeModel) Update(sessionTime *SessionTime) error {
	query := `
        UPDATE session_times
        SET class_id = $1, day = $2, time = $3, duration = $4, version = version + 1
        WHERE id = $5
        RETURNING version
      `
	args := []any{sessionTime.Class_id, sessionTime.Day, sessionTime.Time, sessionTime.Duration, sessionTime.ID}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.DB.QueryRowContext(ctx, query, args...).Scan(&sessionTime.Version)

}

func (c SessionTimeModel) Delete(id int64) error {

	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
        DELETE FROM session_times
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

func (c SessionTimeModel) GetAll(class_id *int, day *string, sessionTime *time.Time, duration *int, filters Filters) ([]*SessionTime, Metadata, error) {

	query := fmt.Sprintf(`
        SELECT  COUNT(*) OVER(), *
        FROM session_times
        WHERE (class_id = $1 OR $1 IS NULL)
            AND (day = $2 OR $2 IS NULL)
            AND ($3::time IS NULL OR time = $3::time)
            AND (duration = $4 OR $4 IS NULL)
        ORDER BY %s %s, id ASC
        LIMIT $5 OFFSET $6`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := c.DB.QueryContext(ctx, query, class_id, day, sessionTime, duration, filters.limit(), filters.offset())

	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()
	totalRecords := 0

	sessionTimes := []*SessionTime{}

	for rows.Next() {
		var sessionTime SessionTime
		err := rows.Scan(&totalRecords, &sessionTime.ID, &sessionTime.Class_id, &sessionTime.Day, &sessionTime.Time, &sessionTime.Duration, &sessionTime.CreatedAt, &sessionTime.Version)

		if err != nil {
			return nil, Metadata{}, err
		}

		sessionTimes = append(sessionTimes, &sessionTime)
	}

	err = rows.Err()
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := CalculateMetaData(totalRecords, filters.Page, filters.PageSize)

	return sessionTimes, metadata, nil
}
