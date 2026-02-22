package main

import (
	//   "encoding/json"
	"fmt"
	"net/http"
	// import the data package which contains the definition for Class
	"github.com/Joseph-Koop/json-project/internal/data"
	"github.com/Joseph-Koop/json-project/internal/validator"
)

func (a *applicationDependencies) postClassHandler(w http.ResponseWriter,
	r *http.Request) {
	// create a struct to hold a class
	// we use struct tags[``] to make the names display in lowercase
	var incomingData struct {
		Studio  string `json:"studio"`
		Trainer  string  `json:"trainer"`
		Capacity_limit  int64 `json:"capacity_limit"`
		Membership_tier  string `json:"membership_tier"`
		Name  string `json:"name"`
		Status  string `json:"status"`
	}
	// perform the decoding
	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	// Copy the values from incomingData to a new Class struct
	// At this point in our code the JSON is well-formed JSON so now
	// we will validate it using the Validator which expects a Class
	class := &data.Class {
		Studio: incomingData.Studio,
		Trainer: incomingData.Trainer,
		Capacity_limit: incomingData.Capacity_limit,
		Membership_tier: incomingData.Membership_tier,
		Name: incomingData.Name,
		Status: incomingData.Status,
	}
	// Initialize a Validator instance
	v := validator.New()

	// Do the validation
	data.ValidateClass(v, class)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)  // implemented later
		return
	}

	// for now display the result
	fmt.Fprintf(w, "%+v\n", incomingData)
}
