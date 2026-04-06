package main

import (
	//   "encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	// import the data package which contains the definition for Class
	"github.com/Joseph-Koop/json-project/internal/data"
	"github.com/Joseph-Koop/json-project/internal/validator"
)

func (a *applicationDependencies) postClassHandler(w http.ResponseWriter,
	r *http.Request) {
	// create a struct to hold a class
	// we use struct tags[``] to make the names display in lowercase
	var incomingData struct {
		Studio_id       int    `json:"studio_id"`
		Trainer_id      int    `json:"trainer_id"`
		Capacity_limit  int    `json:"capacity_limit"`
		Membership_tier string `json:"membership_tier"`
		Name            string `json:"name"`
		Terminated      bool   `json:"terminated"`
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
	class := &data.Class{
		Studio_id:       incomingData.Studio_id,
		Trainer_id:      incomingData.Trainer_id,
		Capacity_limit:  incomingData.Capacity_limit,
		Membership_tier: incomingData.Membership_tier,
		Name:            incomingData.Name,
		Terminated:      incomingData.Terminated,
	}
	// Initialize a Validator instance
	v := validator.New()

	// Do the validation
	data.ValidateClass(v, class)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors) // implemented later
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

func (a *applicationDependencies) displayClassHandler(w http.ResponseWriter, r *http.Request) {
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
	data := envelope{
		"class": class,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

}

func (a *applicationDependencies) updateClassHandler(w http.ResponseWriter, r *http.Request) {
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
		Studio_id       *int    `json:"studio_id"`
		Trainer_id      *int    `json:"trainer_id"`
		Capacity_limit  *int    `json:"capacity_limit"`
		Membership_tier *string `json:"membership_tier"`
		Name            *string `json:"name"`
		Terminated      *bool   `json:"terminated"`
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
	data := envelope{
		"class": class,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}

func (a *applicationDependencies) deleteClassHandler(w http.ResponseWriter, r *http.Request) {
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
	data := envelope{
		"message": "Class successfully deleted.",
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}

}

func (a *applicationDependencies) listClassesHandler(w http.ResponseWriter, r *http.Request) {

	// Create a struct to hold the query parameters
	// Later on we will add fields for pagination and sorting (filters)
	// Joseph: changed to pointers so I can filter by null values in sql function
	var queryParametersData struct {
		Studio_id       *int
		Trainer_id      *int
		Capacity_limit  *int
		Membership_tier *string
		Name            *string
		Terminated      *bool
		data.Filters
	}
	// get the query parameters from the URL
	queryParameters := r.URL.Query()

	// Create a new validator instance
	v := validator.New()

	// Load the query parameters into our struct
	// Joseph: handle conversions
	studio_string := a.getSingleQueryParameter(queryParameters, "studio_id", "")
	if studio_string != "" {
		studio_int, err := strconv.Atoi(studio_string)

		if err == nil && studio_int != 0 {
			queryParametersData.Studio_id = &studio_int
		} else {
			v.AddError(studio_string, "Must be an integer value.")
		}
	}
	trainer_string := a.getSingleQueryParameter(queryParameters, "trainer_id", "")
	if trainer_string != "" {
		trainer_int, err := strconv.Atoi(trainer_string)
		if err == nil && trainer_int != 0 {
			queryParametersData.Trainer_id = &trainer_int
		} else {
			v.AddError(trainer_string, "Must be an integer value.")
		}
	}
	capacity_string := a.getSingleQueryParameter(queryParameters, "capacity_limit", "")
	if capacity_string != "" {
		capacity_int, err := strconv.Atoi(capacity_string)
		if err == nil && capacity_int != 0 {
			queryParametersData.Capacity_limit = &capacity_int
		} else {
			v.AddError(capacity_string, "Must be an integer value.")
		}
	}
	membership_string := a.getSingleQueryParameter(queryParameters, "membership_tier", "")
	if membership_string != "" && (membership_string == "basic" || membership_string == "standard" || membership_string == "premium") {
		queryParametersData.Membership_tier = &membership_string
	}
	name_string := a.getSingleQueryParameter(queryParameters, "name", "")

	if name_string != "" {
		queryParametersData.Name = &name_string
	}
	terminated_string := a.getSingleQueryParameter(queryParameters, "terminated", "")
	if terminated_string != "" {
		terminated_bool, err := strconv.ParseBool(terminated_string)
		if err == nil {
			queryParametersData.Terminated = &terminated_bool
		} else {
			v.AddError(terminated_string, "Must be an boolean value.")
		}
	}

	queryParametersData.Filters.Page = a.getSingleIntegerParameter(queryParameters, "page", 1, v)
	queryParametersData.Filters.PageSize = a.getSingleIntegerParameter(queryParameters, "page_size", 10, v)

	queryParametersData.Filters.Sort = a.getSingleQueryParameter(queryParameters, "sort", "id")
	queryParametersData.Filters.SortSafeList = []string{"id", "capacity_limit", "name", "-id", "-capacity_limit", "-name"}

	// Check if our filters are valid
	data.ValidateFilters(v, queryParametersData.Filters)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	classes, metadata, err := a.classModel.GetAll(queryParametersData.Studio_id, queryParametersData.Trainer_id, queryParametersData.Capacity_limit, queryParametersData.Membership_tier, queryParametersData.Name, queryParametersData.Terminated, queryParametersData.Filters)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{
		"classes":   classes,
		"@metadata": metadata,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
