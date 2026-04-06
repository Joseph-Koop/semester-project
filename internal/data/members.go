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

type Member struct {
	ID              int       `json:"id"`
	User_id         int       `json:"user_id"`
	Name            string    `json:"name"`
	Address         string    `json:"address"`
	Phone           int       `json:"phone"`
	Email           string    `json:"email"`
	Membership_tier string    `json:"membership_tier"`
	Expiry_date     time.Time `json:"expiry_date"`
	CreatedAt       time.Time `json:"-"`
	Version         int32     `json:"version"`
}

func ValidateMember(v *validator.Validator, member *Member) {

	v.Check(len(strconv.Itoa(member.User_id)) > 0, "user_id", "Must be an existing user.")

	v.Check(member.Name != "", "name", "Must be provided.")
	v.Check(len(member.Name) <= 50, "name", "Must not be more than 50 bytes long.")

	v.Check(member.Address != "", "address", "Must be provided.")
	v.Check(len(member.Address) <= 50, "address", "Must not be more than 100 bytes long.")

	v.Check(len(strconv.Itoa(member.Phone)) == 10, "phone", "Must be 10 digits long.")

	v.Check(member.Email != "", "email", "Must be provided.")
	v.Check(len(member.Email) <= 50, "email", "Must not be more than 50 bytes long.")
	_, err := mail.ParseAddress(member.Email)
	v.Check(err == nil, "email", "Must be a valid email address")

	v.Check(member.Membership_tier == "basic" || member.Membership_tier == "standard" || member.Membership_tier == "premium", "membership_tier", "Must be one of the valid options.")

	v.Check(!member.Expiry_date.IsZero(), "expiry_date", "Must be provided.")
	today := time.Now().Truncate(24 * time.Hour)
	v.Check(member.Expiry_date.After(today), "expiry_date", "Must be today or later.")
}

type MemberModel struct {
	DB *sql.DB
}

func (c MemberModel) Insert(member *Member) error {

	query := `
        INSERT INTO members (user_id, name, address, phone, email, membership_tier, expiry_date)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, created_at, version
        `

	args := []any{member.User_id, member.Name, member.Address, member.Phone, member.Email, member.Membership_tier, member.Expiry_date}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.DB.QueryRowContext(ctx, query, args...).Scan(&member.ID, &member.CreatedAt, &member.Version)

}

func (c MemberModel) Get(id int) (*Member, error) {

	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
        SELECT *
        FROM members
        WHERE id = $1
      `

	var member Member

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, id).Scan(&member.ID, &member.User_id, &member.Name, &member.Address, &member.Phone, &member.Email, &member.Membership_tier, &member.Expiry_date, &member.CreatedAt, &member.Version)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &member, nil
}

func (c MemberModel) Update(member *Member) error {
	query := `
        UPDATE members
        SET user_id = $1, name = $2, address = $3, phone = $4, email = $5, membership_tier = $6, expiry_date = $7, version = version + 1
        WHERE id = $8
        RETURNING version
      `
	args := []any{member.User_id, member.Name, member.Address, member.Phone, member.Email, member.Membership_tier, member.Expiry_date, member.ID}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.DB.QueryRowContext(ctx, query, args...).Scan(&member.Version)

}

func (c MemberModel) Delete(id int) error {

	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
        DELETE FROM members
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

func (c MemberModel) GetAll(user_id *int, name *string, address *string, phone *int, email *string, membership_tier *string, expiry_date *time.Time, filters Filters) ([]*Member, Metadata, error) {

	query := fmt.Sprintf(`
        SELECT  COUNT(*) OVER(), *
        FROM members
        WHERE (to_tsvector('simple', name) @@ 
                plainto_tsquery('simple', $1) OR $1 IS NULL)
            AND (to_tsvector('simple', address) @@ 
                plainto_tsquery('simple', $2) OR $2 IS NULL)
            AND (phone = $3 OR $3 IS NULL)
            AND (to_tsvector('simple', email) @@ 
                plainto_tsquery('simple', $4) OR $4 IS NULL)
            AND (membership_tier = $5 OR $5 IS NULL)
            AND ($6::date IS NULL OR expiry_date = $6::date)
        ORDER BY %s %s, id ASC
        LIMIT $7 OFFSET $8`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := c.DB.QueryContext(ctx, query, name, address, phone, email, membership_tier, expiry_date, filters.limit(), filters.offset())

	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()
	totalRecords := 0

	members := []*Member{}

	for rows.Next() {
		var member Member
		err := rows.Scan(&totalRecords, &member.ID, &member.User_id, &member.Name, &member.Address, &member.Phone, &member.Email, &member.Membership_tier, &member.Expiry_date, &member.CreatedAt, &member.Version)

		if err != nil {
			return nil, Metadata{}, err
		}

		members = append(members, &member)
	}

	err = rows.Err()
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := CalculateMetaData(totalRecords, filters.Page, filters.PageSize)

	return members, metadata, nil
}
