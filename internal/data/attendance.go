package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Joseph-Koop/json-project/internal/validator"
)

type Attendance struct {
	ID              int       `json:"id"`
	Registration_id int       `json:"registration_id"`
	Session_id      int       `json:"session_id"`
	CreatedAt       time.Time `json:"-"`
	Version         int32     `json:"version"`
}

type AttendanceModel struct {
	DB *sql.DB
}

func (c AttendanceModel) ValidateAttendance(v *validator.Validator, attendance *Attendance) {

	v.Check(attendance.Registration_id > 0, "registration_id", "Must be an existing registration.")

	v.Check(attendance.Session_id > 0, "session_id", "Must be an existing session.")

}

func (c AttendanceModel) Insert(attendance *Attendance) error {

	query := `
        INSERT INTO attendance (registration_id, session_id)
        VALUES ($1, $2)
        RETURNING id, created_at, version
        `

	args := []any{attendance.Registration_id, attendance.Session_id}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.DB.QueryRowContext(ctx, query, args...).Scan(&attendance.ID, &attendance.CreatedAt, &attendance.Version)

}

func (c AttendanceModel) Get(id int) (*Attendance, error) {

	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
        SELECT *
        FROM attendance
        WHERE id = $1
      `

	var attendance Attendance

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, id).Scan(&attendance.ID, &attendance.Registration_id, &attendance.Session_id, &attendance.CreatedAt, &attendance.Version)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &attendance, nil
}

func (c AttendanceModel) Update(attendance *Attendance) error {
	query := `
        UPDATE attendance
        SET registration_id = $1, session_id = $2, version = version + 1
        WHERE id = $3
        RETURNING version
      `
	args := []any{attendance.Registration_id, attendance.Session_id, attendance.ID}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.DB.QueryRowContext(ctx, query, args...).Scan(&attendance.Version)

}

func (c AttendanceModel) Delete(id int) error {

	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
        DELETE FROM attendance
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

func (c AttendanceModel) GetAll(registration_id *int, session_id *int, filters Filters) ([]*Attendance, Metadata, error) {

	query := fmt.Sprintf(`
        SELECT  COUNT(*) OVER(), *
        FROM attendance
        WHERE (registration_id = $1 OR $1 IS NULL)
		AND (session_id = $2 OR $2 IS NULL)
        ORDER BY %s %s, id ASC
        LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := c.DB.QueryContext(ctx, query, registration_id, session_id, filters.limit(), filters.offset())

	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()
	totalRecords := 0

	attendances := []*Attendance{}

	for rows.Next() {
		var attendance Attendance
		err := rows.Scan(&totalRecords, &attendance.ID, &attendance.Registration_id, &attendance.Session_id, &attendance.CreatedAt, &attendance.Version)

		if err != nil {
			return nil, Metadata{}, err
		}

		attendances = append(attendances, &attendance)
	}

	err = rows.Err()
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := CalculateMetaData(totalRecords, filters.Page, filters.PageSize)

	return attendances, metadata, nil
}

func (c AttendanceModel) GetAllByMemberID(member_id int, registration_id *int, session_id *int, filters Filters) ([]*Attendance, Metadata, error) {

	query := fmt.Sprintf(`
        SELECT  COUNT(*) OVER(), a.*
        FROM attendance a
		INNER JOIN registrations r ON a.registration_id = r.id
        WHERE r.member_id = $1
		AND (a.registration_id = $2 OR $2 IS NULL)
		AND (a.session_id = $3 OR $3 IS NULL)
        ORDER BY %s %s, a.id ASC
        LIMIT $4 OFFSET $5`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := c.DB.QueryContext(ctx, query, member_id, registration_id, session_id, filters.limit(), filters.offset())

	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()
	totalRecords := 0

	attendances := []*Attendance{}

	for rows.Next() {
		var attendance Attendance
		err := rows.Scan(&totalRecords, &attendance.ID, &attendance.Registration_id, &attendance.Session_id, &attendance.CreatedAt, &attendance.Version)

		if err != nil {
			return nil, Metadata{}, err
		}

		attendances = append(attendances, &attendance)
	}

	err = rows.Err()
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := CalculateMetaData(totalRecords, filters.Page, filters.PageSize)

	return attendances, metadata, nil
}

func (c AttendanceModel) GetAllByTrainerID(trainer_id int, registration_id *int, session_id *int, filters Filters) ([]*Attendance, Metadata, error) {

	query := fmt.Sprintf(`
        SELECT  COUNT(*) OVER(), a.*
        FROM attendance a
		INNER JOIN sessions s ON s.id = a.session_id
		INNER JOIN classes c ON c.id = s.class_id
		WHERE c.trainer_id = $1
        AND (a.registration_id = $2 OR $2 IS NULL)
		AND (a.session_id = $3 OR $3 IS NULL)
        ORDER BY %s %s, a.id ASC
        LIMIT $4 OFFSET $5`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := c.DB.QueryContext(ctx, query, trainer_id, registration_id, session_id, filters.limit(), filters.offset())

	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()
	totalRecords := 0

	attendances := []*Attendance{}

	for rows.Next() {
		var attendance Attendance
		err := rows.Scan(&totalRecords, &attendance.ID, &attendance.Registration_id, &attendance.Session_id, &attendance.CreatedAt, &attendance.Version)

		if err != nil {
			return nil, Metadata{}, err
		}

		attendances = append(attendances, &attendance)
	}

	err = rows.Err()
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := CalculateMetaData(totalRecords, filters.Page, filters.PageSize)

	return attendances, metadata, nil
}
