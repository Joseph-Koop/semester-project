package data

import (
  "time"
  "github.com/Joseph-Koop/json-project/internal/validator"
)

// each name begins with uppercase so that they are exportable/public
type Class struct {
    ID int64                  `json:"id"`
    Studio  string            `json:"studio"`
    Trainer  string           `json:"trainer"`
    Capacity_limit  int64     `json:"capacity_limit"`
    Membership_tier  string   `json:"membership_tier"`
    Name  string              `json:"name"`
    Status  string            `json:"status"`
    CreatedAt  time.Time      `json:"-"`
    //Version?
} 

// Create a function that performs the validation checks
func ValidateClass(v *validator.Validator, class *Class) {
    v.Check(class.Studio != "", "studio", "Must be provided.")                                  // check if the Studio field is empty
    v.Check(class.Trainer != "", "trainer", "Must be provided.")                                // check if the Trainer field is empty
    v.Check(class.Capacity_limit > 0, "capacity_limit", "Must be greater than 0.")              // check if the Capacity_limit field 0 or less
    v.Check(class.Membership_tier != "", "membership_tier", "Must be provided.")                // check if the Membership_tier field is empty
    v.Check(class.Name != "", "name", "Must be provided.")                                      // check if the Name field is empty
    v.Check(class.Status != "", "status", "Must be provided.")                                  // check if the Status field is empty

    v.Check(len(class.Studio) <= 50, "studio", "Must not be more than 50 bytes long.")                      // check if the Studio field is bigger than 50 characters
    v.Check(len(class.Trainer) <= 50, "trainer", "Must not be more than 50 bytes long.")                    // check if the Trainer field is bigger than 50 characters
    v.Check(class.Capacity_limit <= 100, "capacity_limit", "Must be less than or equal to 100.")            // check if the Capacity_limit field is bigger than 100
    v.Check(len(class.Membership_tier) <= 50, "membership_tier", "Must not be more than 50 bytes long.")    // check if the Membership_tier field is bigger than 50 characters
    v.Check(len(class.Name) <= 50, "name", "Must not be more than 50 bytes long.")                          // check if the Name field is bigger than 50 characters
    v.Check(len(class.Status) <= 50, "status", "Must not be more than 50 bytes long.")                      // check if the Status field is bigger than 50 characters
}
