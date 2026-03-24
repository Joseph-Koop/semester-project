package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Joseph-Koop/json-project/internal/validator"
)

type Registration struct {
	ID        int     `json:"id"`
	Class_id      int    `json:"class_id"`
	Member_id      int    `json:"member_id"`
	Status      string    `json:"status"`
	CreatedAt time.Time `json:"-"`
	Version   int32     `json:"version"`
}

type RegistrationModel struct {
	DB *sql.DB
}

func (c RegistrationModel) ValidateRegistration(v *validator.Validator, registration *Registration) {

	v.Check(registration.Class_id > 0, "class_id", "Must be an existing class.")

	v.Check(registration.Member_id > 0, "member_id", "Must be an existing member.")

	v.Check(registration.Status == "active" || registration.Status == "dropped", "status", "Must be one of the valid options.")

	var member_membership_tier string
	var member_expiry_date time.Time
	query := `
		SELECT membership_tier, expiry_date FROM MEMBERS WHERE id = $1
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := c.DB.QueryRowContext(ctx, query, registration.Member_id).Scan(&member_membership_tier, &member_expiry_date)
	if err != nil {
		v.AddError("member_id", "Internal database operation failed.")
		return
	}
	
	var class_membership_tier string
	var class_capacity_limit int
	var class_terminated bool
	query2 := `
	SELECT membership_tier, capacity_limit, terminated FROM CLASSES WHERE id = $1
	`
	ctx2, cancel2 := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel2()
	err = c.DB.QueryRowContext(ctx2, query2, registration.Class_id).Scan(&class_membership_tier, &class_capacity_limit, &class_terminated)
	if err != nil {
		v.AddError("class_id", "Class not found.")
		return
	}
	
	hierarchy := map[string]int{
		"basic": 1,
		"standard": 2,
		"premium": 3,
	}

	var current_class_quantity int
	query3 := `
		SELECT COUNT(*) FROM registrations WHERE status = 'active' AND class_id = $1
	`
	ctx3, cancel3 := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel3()
	err = c.DB.QueryRowContext(ctx3, query3, registration.Class_id).Scan(&current_class_quantity)
	if err != nil {
		v.AddError("member_id", "Internal count operation failed.")
		return
	}



	conflictQuery := `
		SELECT 1
		FROM registrations r
		JOIN session_times s_existing 
			ON r.class_id = s_existing.class_id
		JOIN session_times s_new
			ON s_new.class_id = $2
		WHERE r.member_id = $1
		AND r.status = 'active'
		AND s_existing.day = s_new.day
		AND s_existing.time < (s_new.time + (s_new.duration || ' minutes')::interval)
		AND s_new.time < (s_existing.time + (s_existing.duration || ' minutes')::interval)
		LIMIT 1;
	`

	var exists int

	ctx4, cancel4 := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel4()

	err = c.DB.QueryRowContext(ctx4, conflictQuery,
		registration.Member_id,
		registration.Class_id,
	).Scan(&exists)

	if err == nil {
		v.AddError("class_id", "Schedule conflict with existing class.")
		return
	}

	if err != sql.ErrNoRows {
		v.AddError("member_id", "Internal database operation failed.")
		return
	}





	if class_terminated == true{
		v.AddError("class_id", "This class is no longer active.")
	}

	if(current_class_quantity >= class_capacity_limit){
		v.AddError("class_id", "The class capacity is full.")
	}

	if(hierarchy[member_membership_tier] < hierarchy[class_membership_tier]){
		v.AddError("member_id", "Must have a sufficient membership.")
	}

	if member_expiry_date.Before(time.Now()){
		v.AddError("member_id", "Must have an active membership.")
	}
}

func (c RegistrationModel) Insert(registration *Registration) error {

	query := `
        INSERT INTO registrations (class_id, member_id, status)
        VALUES ($1, $2, $3)
        RETURNING id, created_at, version
        `

	args := []any{registration.Class_id, registration.Member_id, registration.Status}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.DB.QueryRowContext(ctx, query, args...).Scan(&registration.ID, &registration.CreatedAt, &registration.Version)

}

func (c RegistrationModel) Get(id int64) (*Registration, error) {

	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
        SELECT *
        FROM registrations
        WHERE id = $1
      `

	var registration Registration

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, id).Scan(&registration.ID, &registration.Class_id, &registration.Member_id, &registration.Status, &registration.CreatedAt, &registration.Version)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &registration, nil
}

func (c RegistrationModel) Update(registration *Registration) error {
	query := `
        UPDATE registrations
        SET class_id = $1, member_id = $2, status = $3, version = version + 1
        WHERE id = $4
        RETURNING version
      `
	args := []any{registration.Class_id, registration.Member_id, registration.Status, registration.ID}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.DB.QueryRowContext(ctx, query, args...).Scan(&registration.Version)

}

func (c RegistrationModel) Delete(id int64) error {

	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
        DELETE FROM registrations
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

func (c RegistrationModel) GetAll(class_id *int, member_id *int, status *string, filters Filters) ([]*Registration, Metadata, error) {

	query := fmt.Sprintf(`
        SELECT  COUNT(*) OVER(), *
        FROM registrations
        WHERE (class_id = $1 OR $1 IS NULL)
		AND (member_id = $2 OR $2 IS NULL)
		AND (status = $3 OR $3 IS NULL)
        ORDER BY %s %s, id ASC
        LIMIT $4 OFFSET $5`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := c.DB.QueryContext(ctx, query, class_id, member_id, status, filters.limit(), filters.offset())

	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()
	totalRecords := 0

	registrations := []*Registration{}

	for rows.Next() {
		var registration Registration
		err := rows.Scan(&totalRecords, &registration.ID, &registration.Class_id, &registration.Member_id, &registration.Status, &registration.CreatedAt, &registration.Version)

		if err != nil {
			return nil, Metadata{}, err
		}

		registrations = append(registrations, &registration)
	}

	err = rows.Err()
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := CalculateMetaData(totalRecords, filters.Page, filters.PageSize)

	return registrations, metadata, nil
}
