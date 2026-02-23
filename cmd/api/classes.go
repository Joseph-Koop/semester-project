package main

import (
	//   "encoding/json"
	"errors"
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
		Studio_id  int64 `json:"studio_id"`
		Trainer_id  int64  `json:"trainer_id"`
		Capacity_limit  int64 `json:"capacity_limit"`
		Membership_tier  string `json:"membership_tier"`
		Name  string `json:"name"`
		Terminated  bool `json:"terminated"`
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
		Studio_id: incomingData.Studio_id,
		Trainer_id: incomingData.Trainer_id,
		Capacity_limit: incomingData.Capacity_limit,
		Membership_tier: incomingData.Membership_tier,
		Name: incomingData.Name,
		Terminated: incomingData.Terminated,
	}
	// Initialize a Validator instance
	v := validator.New()

	// Do the validation
	data.ValidateClass(v, class)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)  // implemented later
		return
	}

	// Add the class to the database table
   err = a.classModel.Insert(class)
   if err != nil {
       a.serverErrorResponse(w, r, err)
       return
   }

	// fmt.Fprintf(w, "%+v\n", incomingData)      // delete this

   // Set a Location header. The path to the newly created class
   headers := make(http.Header)
   headers.Set("Location", fmt.Sprintf("/classes/%d", class.ID))

	// Send a JSON response with 201 (new resource created) status code
	data := envelope{
			"class": class,
		}
	err = a.writeJSON(w, http.StatusCreated, data, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}

func (a *applicationDependencies)displayClassHandler(w http.ResponseWriter, r *http.Request) {
	// Get the id from the URL /v1/comments/:id so that we
	// can use it to query teh comments table. We will 
	// implement the readIDParam() function later
   id, err := a.readIDParam(r)
   if err != nil {
       a.notFoundResponse(w, r)
       return 
   }

   // Call Get() to retrieve the class with the specified id
   class, err := a.classModel.Get(id)
   if err != nil {
       switch {
           case errors.Is(err, data.ErrRecordNotFound):
              a.notFoundResponse(w, r)
           default:
              a.serverErrorResponse(w, r, err)
       }
       return 
   }

   // display the class
    data := envelope {
		"class": class,
	}
    err = a.writeJSON(w, http.StatusOK, data, nil)
    if err != nil {
       a.serverErrorResponse(w, r, err)
       return 
   }

}


