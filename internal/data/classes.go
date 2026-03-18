package data

import (
  "context"
  "database/sql"
  "errors"
  "fmt"
  "time"
  "github.com/Joseph-Koop/json-project/internal/validator"
)

// each name begins with uppercase so that they are exportable/public
type Class struct {
    ID int64                  `json:"id"`
    Studio_id  int64            `json:"studio_id"`
    Trainer_id  int64           `json:"trainer_id"`
    Capacity_limit  int64     `json:"capacity_limit"`
    Membership_tier  string   `json:"membership_tier"`
    Name  string              `json:"name"`
    Terminated  bool            `json:"terminated"`
    CreatedAt  time.Time      `json:"-"`
    Version int32             `json:"version"`
} 

// Create a function that performs the validation checks
func ValidateClass(v *validator.Validator, class *Class) {
    v.Check(class.Studio_id > 0, "studio_id", "Must be an existing studio.")                                  // check if the Studio field is 0 or less

    v.Check(class.Trainer_id > 0, "trainer_id", "Must be an existing trainer.")                                // check if the Trainer field is 0 or less

    v.Check(class.Capacity_limit > 0, "capacity_limit", "Must be greater than 0.")                    // check if the Capacity_limit field 0 or less
    v.Check(class.Capacity_limit <= 100, "capacity_limit", "Must be less than or equal to 100.")      // check if the Capacity_limit field is bigger than 100

    v.Check(class.Membership_tier == "basic" || class.Membership_tier == "standard" || class.Membership_tier == "premium", "membership_tier", "Must be one of the valid options.")                // check if the Membership_tier field matches one of the options
    
    v.Check(class.Name != "", "name", "Must be provided.")                                      // check if the Name field is empty
    v.Check(len(class.Name) <= 50, "name", "Must not be more than 50 bytes long.")              // check if the Name field is bigger than 50 characters

    v.Check(class.Terminated == false || class.Terminated == true, "terminated", "Must be provided.")       // check if the Status field is either true of ralse

}

// A ClassModel expects a connection pool
type ClassModel struct {
    DB *sql.DB
}

// Insert a new row in the class table
// Expects a pointer to the actual class
func (c ClassModel) Insert(class *Class) error {
   // the SQL query to be executed against the database table
    query := `
        INSERT INTO classes (studio_id, trainer_id, capacity_limit, membership_tier, name, terminated)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, created_at, version
        `
  // the actual values to replace $1, and $2
   args := []any{class.Studio_id, class.Trainer_id, class.Capacity_limit, class.Membership_tier, class.Name, class.Terminated}
   // Create a context with a 3-second timeout. No database
// operation should take more than 3 seconds or we will quit it
   ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
   defer cancel()
// execute the query against the comments database table. We ask for the the
// id, created_at, and version to be sent back to us which we will use
// to update the Comment struct later on 
   return c.DB.QueryRowContext(ctx, query, args...).Scan(&class.ID, &class.CreatedAt, &class.Version)

}


// Get a specific Comment from the comments table
func (c ClassModel) Get(id int64) (*Class, error) {
   // check if the id is valid
    if id < 1 {
        return nil, ErrRecordNotFound
    }
   // the SQL query to be executed against the database table
    query := `
        SELECT *
        FROM classes
        WHERE id = $1
      `
	// declare a variable of type Class to store the returned class
   var class Class

	// Set a 3-second context/timer
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, id).Scan (&class.ID, &class.Studio_id, &class.Trainer_id, &class.Capacity_limit, &class.Membership_tier, &class.Name, &class.Terminated, &class.CreatedAt, &class.Version,)

	// check for which type of error
	if err != nil {
		switch {
			case errors.Is(err, sql.ErrNoRows):
				return nil, ErrRecordNotFound
			default:
				return nil, err
			}
		}
	return &class, nil
}

// Update a specific Comment from the comments table
func (c ClassModel) Update(class *Class) error {
// The SQL query to be executed against the database table
// Every time we make an update, we increment the version number
    query := `
        UPDATE classes
        SET studio_id = $1, trainer_id = $2, capacity_limit = $3, membership_tier = $4, name = $5, terminated = $6, version = version + 1
        WHERE id = $7
        RETURNING version
      `
    args := []any{class.Studio_id, class.Trainer_id, class.Capacity_limit, class.Membership_tier, class.Name, class.Terminated, class.ID}
    ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
    defer cancel()

    return c.DB.QueryRowContext(ctx, query, args...).Scan(&class.Version)
                                              
}

// Delete a specific Comment from the comments table
func (c ClassModel) Delete(id int64) error {

    // check if the id is valid
    if id < 1 {
        return ErrRecordNotFound
    }
    // the SQL query to be executed against the database table
    query := `
        DELETE FROM classes
        WHERE id = $1
      `
    ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
    defer cancel()

    // ExecContext does not return any rows unlike QueryRowContext. 
    // It only returns  information about the the query execution
    // such as how many rows were affected
    result, err := c.DB.ExecContext(ctx, query, id)
    if err != nil {
        return err
    }

    // Were any rows deleted?
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }
    // Probably a wrong id was provided or the client is trying to
    // delete an already deleted comment
    if rowsAffected == 0 {
        return ErrRecordNotFound
    }

    return nil

}

// Get all classes
func (c ClassModel) GetAll(studio_id *int, trainer_id *int, capacity_limit *int, membership_tier *string, name *string, terminated *bool, filters Filters) ([]*Class, Metadata, error) {

// the SQL query to be executed against the database table
    query := fmt.Sprintf(`
        SELECT  COUNT(*) OVER(), *
        FROM classes
        WHERE (studio_id = $1 OR $1 IS NULL)
            AND (trainer_id = $2 OR $2 IS NULL)
            AND (capacity_limit = $3 OR $3 IS NULL)
            AND (membership_tier = $4 OR $4 IS NULL)
            AND (to_tsvector('simple', name) @@ 
                plainto_tsquery('simple', $5) OR $5 IS NULL) 
            AND (terminated = $6 OR $6 IS NULL)
        ORDER BY %s %s, id ASC
        LIMIT $7 OFFSET $8`, filters.sortColumn(), filters.sortDirection())

    ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
    defer cancel()

    // QueryContext returns multiple rows.
    rows, err := c.DB.QueryContext(ctx, query, studio_id, trainer_id, capacity_limit, membership_tier, name, terminated, filters.limit(), filters.offset())

    if err != nil {
        return nil, Metadata{}, err
    }

    // clean up the memory that was used
    defer rows.Close()
    totalRecords := 0
    // we will store the address of each comment in our slice
    classes := []*Class{}

    // process each row that is in rows

    for rows.Next() {
        var class Class
        err := rows.Scan(&totalRecords, &class.ID, &class.Studio_id, &class.Trainer_id, &class.Capacity_limit, &class.Membership_tier, &class.Name, &class.Terminated, &class.CreatedAt, &class.Version,)

        if err != nil {
            return nil, Metadata{}, err
        }

        // add the row to our slice
        classes = append(classes, &class)
    }  // end of for loop

    // after we exit the loop we need to check if it generated any errors
    err = rows.Err()
    if err != nil {
        return nil, Metadata{}, err
    }
    // Create the metadata
    metadata := CalculateMetaData(totalRecords, filters.Page, filters.PageSize)


    return classes, metadata, nil

}

