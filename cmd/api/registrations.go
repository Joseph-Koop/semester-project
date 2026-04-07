package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Joseph-Koop/json-project/internal/data"
	"github.com/Joseph-Koop/json-project/internal/validator"
)

func (a *applicationDependencies) postRegistrationHandler(w http.ResponseWriter, r *http.Request) {
	var incomingData struct {
		Member_id             int `json:"member_id"`
		Class_id          int `json:"class_id"`
		Status            string `json:"status"`
	}
	
	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}
	
	registration := &data.Registration{
		Member_id: 		incomingData.Member_id,
		Class_id: 		incomingData.Class_id,
		Status: 	incomingData.Status,
	}
	
	v := validator.New()
	
	a.registrationModel.ValidateRegistration(v, registration)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors) 
		return
	}
	
	err = a.registrationModel.Insert(registration)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/registrations/%d", registration.ID))

	data := envelope{
		"registration": registration,
	}
	err = a.writeJSON(w, http.StatusCreated, data, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}

func (a *applicationDependencies) displayRegistrationHandler(w http.ResponseWriter, r *http.Request) {
	
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	registration, err := a.registrationModel.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	data := envelope{
		"registration": registration,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}

func (a *applicationDependencies) updateRegistrationHandler(w http.ResponseWriter, r *http.Request) {
	
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	registration, err := a.registrationModel.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}
	
	var incomingData struct {
		Member_id      	*int `json:"member_id"`
		Class_id      	*int `json:"class_id"`
		Status     *string `json:"status"`
	}

	
	err = a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}
	
	if incomingData.Member_id != nil {
		registration.Member_id = *incomingData.Member_id
	}
	
	if incomingData.Class_id != nil {
		registration.Class_id = *incomingData.Class_id
	}
	
	if incomingData.Status != nil {
		registration.Status = *incomingData.Status
	}
	
	v := validator.New()
	a.registrationModel.ValidateRegistration(v, registration)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	
	err = a.registrationModel.Update(registration)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	data := envelope{
		"registration": registration,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}

func (a *applicationDependencies) deleteRegistrationHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	err = a.registrationModel.Delete(id)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	
	data := envelope{
		"message": "Registration successfully deleted.",
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}

}

func (a *applicationDependencies) listRegistrationsHandler(w http.ResponseWriter, r *http.Request) {
	
	var queryParametersData struct {
		Class_id     	*int
		Member_id     	*int
		Status   	*string
		data.Filters
	}
	
	queryParameters := r.URL.Query()

	v := validator.New()

	class_id_string := a.getSingleQueryParameter(queryParameters, "class_id", "")
	if class_id_string != "" {
		class_id_int, err := strconv.Atoi(class_id_string)

		if err == nil && class_id_int != 0 {
			queryParametersData.Class_id = &class_id_int
		} else {
			v.AddError(class_id_string, "Must be an integer value.")
		}
	}

	member_id_string := a.getSingleQueryParameter(queryParameters, "member_id", "")
	if member_id_string != "" {
		class_id_int, err := strconv.Atoi(member_id_string)

		if err == nil && class_id_int != 0 {
			queryParametersData.Member_id = &class_id_int
		} else {
			v.AddError(member_id_string, "Must be an integer value.")
		}
	}

	status_string := a.getSingleQueryParameter(queryParameters, "status", "")
	if status_string != "" {
		queryParametersData.Status = &status_string
	}

	queryParametersData.Filters.Page = a.getSingleIntegerParameter(queryParameters, "page", 1, v)
	queryParametersData.Filters.PageSize = a.getSingleIntegerParameter(queryParameters, "page_size", 10, v)

	queryParametersData.Filters.Sort = a.getSingleQueryParameter(queryParameters, "sort", "id")
	queryParametersData.Filters.SortSafeList = []string{"id", "class_id", "member_id", "status", "-id", "-class_id", "-member_id", "-status"}

	
	data.ValidateFilters(v, queryParametersData.Filters)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	user := a.contextGetUser(r)

	var registrations []*data.Registration
	var err error
	var metadata any

	switch user.Role_id{
		case 1:
			registrations, metadata, err = a.registrationModel.GetAll(queryParametersData.Class_id, queryParametersData.Member_id, queryParametersData.Status, queryParametersData.Filters)


		case 2:
			registrations, metadata, err = a.registrationModel.GetAll(queryParametersData.Class_id, queryParametersData.Member_id, queryParametersData.Status, queryParametersData.Filters)

		case 3:
			member, err2 := a.memberModel.GetByUserID(user.ID)
			if err2 != nil {
				a.serverErrorResponse(w, r, err2)
				return
			}

			registrations, metadata, err = a.registrationModel.GetAllByMemberID(member.ID, queryParametersData.Class_id, queryParametersData.Member_id, queryParametersData.Status, queryParametersData.Filters)
		default:
			a.notPermittedResponse(w, r)
			return
	}

	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{
		"registrations":   registrations,
		"@metadata": metadata,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
