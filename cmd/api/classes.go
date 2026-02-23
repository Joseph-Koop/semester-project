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

func (a *applicationDependencies)updateClassHandler(w http.ResponseWriter, r *http.Request) {
	// Get the id from the URL
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return 
	}

	// Call Get() to retrieve the comment with the specified id
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

	// Use our temporary incomingData struct to hold the data
	// Note: I have changed the types to pointer to differentiate
	// between the client leaving a field empty intentionally
	// and the field not needing to be updated
	var incomingData struct {
		Studio_id *int64 `json:"studio_id"`
		Trainer_id *int64  `json:"trainer_id"`
		Capacity_limit *int64 `json:"capacity_limit"`
		Membership_tier *string `json:"membership_tier"`
		Name *string `json:"name"`
		Terminated *bool `json:"terminated"`
    }

	// perform the decoding
	err = a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}
	// We need to now check the fields to see which ones need updating
	// if incomingData.Studio_id is nil, no update was provided
	if incomingData.Studio_id != nil {
		class.Studio_id = *incomingData.Studio_id
	}
	// if incomingData.Trainer_id is nil, no update was provided
	if incomingData.Trainer_id != nil {
		class.Trainer_id = *incomingData.Trainer_id
	}
	// if incomingData.Capacity_limit is nil, no update was provided
	if incomingData.Capacity_limit != nil {
		class.Capacity_limit = *incomingData.Capacity_limit
	}
	// if incomingData.Membership_tier is nil, no update was provided
	if incomingData.Membership_tier != nil {
		class.Membership_tier = *incomingData.Membership_tier
	}
	// if incomingData.Name is nil, no update was provided
	if incomingData.Name != nil {
		class.Name = *incomingData.Name
	}
	// if incomingData.Terminated is nil, no update was provided
	if incomingData.Terminated != nil {
		class.Terminated = *incomingData.Terminated
	}

	// Before we write the updates to the DB let's validate
	v := validator.New()
	data.ValidateClass(v, class)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)  
		return
	}

	// perform the update
    err = a.classModel.Update(class)
    if err != nil {
		a.serverErrorResponse(w, r, err)
		return 
	}
	data := envelope {
		"class": class,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return 
	}
}


func (a *applicationDependencies)deleteClassHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return 
	}

	err = a.classModel.Delete(id)

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
		"message": "Class successfully deleted.",
	}
    err = a.writeJSON(w, http.StatusOK, data, nil)
    if err != nil {
       a.serverErrorResponse(w, r, err)
    }

}





