package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Joseph-Koop/json-project/internal/data"
	"github.com/Joseph-Koop/json-project/internal/validator"
)

func (a *applicationDependencies) postTrainerHandler(w http.ResponseWriter, r *http.Request) {
	var incomingData struct {
		User_id            int `json:"user_id"`
		Name             string `json:"name"`
		Address          string `json:"address"`
		Phone            int `json:"phone"`
		Email            string `json:"email"`
	}
	
	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}
	
	trainer := &data.Trainer{
		User_id: 		incomingData.User_id,
		Name: 		incomingData.Name,
		Address: 	incomingData.Address,
		Phone: 		incomingData.Phone,
		Email: 		incomingData.Email,
	}
	
	v := validator.New()
	
	data.ValidateTrainer(v, trainer)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors) 
		return
	}
	
	err = a.trainerModel.Insert(trainer)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/trainers/%d", trainer.ID))

	data := envelope{
		"trainer": trainer,
	}
	err = a.writeJSON(w, http.StatusCreated, data, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}

func (a *applicationDependencies) displayTrainerHandler(w http.ResponseWriter, r *http.Request) {
	
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	trainer, err := a.trainerModel.Get(id)
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
		"trainer": trainer,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}

func (a *applicationDependencies) updateTrainerHandler(w http.ResponseWriter, r *http.Request) {
	
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	trainer, err := a.trainerModel.Get(id)
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
		User_id      	*int `json:"user_id"`
		Name      	*string `json:"name"`
		Address     *string `json:"address"`
		Phone      	*int `json:"phone"`
		Email      	*string `json:"email"`
	}

	
	err = a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}
	
	if incomingData.User_id != nil {
		trainer.User_id = *incomingData.User_id
	}
	
	if incomingData.Name != nil {
		trainer.Name = *incomingData.Name
	}
	
	if incomingData.Address != nil {
		trainer.Address = *incomingData.Address
	}
	
	if incomingData.Phone != nil {
		trainer.Phone = *incomingData.Phone
	}
	
	if incomingData.Email != nil {
		trainer.Email = *incomingData.Email
	}
	
	v := validator.New()
	data.ValidateTrainer(v, trainer)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	
	err = a.trainerModel.Update(trainer)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	data := envelope{
		"trainer": trainer,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}

func (a *applicationDependencies) deleteTrainerHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	err = a.trainerModel.Delete(id)

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
		"message": "Trainer successfully deleted.",
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}

}

func (a *applicationDependencies) listTrainersHandler(w http.ResponseWriter, r *http.Request) {
	
	var queryParametersData struct {
		User_id     	*int
		Name     	*string
		Address   	*string
		Phone     	*int
		Email     	*string
		data.Filters
	}
	
	queryParameters := r.URL.Query()

	v := validator.New()

	user_id_string := a.getSingleQueryParameter(queryParameters, "user_id", "")
	if user_id_string != "" {
		user_id_int, err := strconv.Atoi(user_id_string)

		if err == nil && user_id_int != 0 {
			queryParametersData.Phone = &user_id_int
		} else {
			v.AddError(user_id_string, "Must be an integer value.")
		}
	}

	name_string := a.getSingleQueryParameter(queryParameters, "name", "")
	if name_string != "" {
		queryParametersData.Name = &name_string
	}

	address_string := a.getSingleQueryParameter(queryParameters, "address", "")
	if address_string != "" {
		queryParametersData.Address = &address_string
	}

	phone_string := a.getSingleQueryParameter(queryParameters, "phone", "")
	if phone_string != "" {
		phone_int, err := strconv.Atoi(phone_string)

		if err == nil && phone_int != 0 {
			queryParametersData.Phone = &phone_int
		} else {
			v.AddError(phone_string, "Must be an integer value.")
		}
	}

	email_string := a.getSingleQueryParameter(queryParameters, "email", "")
	if email_string != "" {
		queryParametersData.Email = &email_string
	}

	queryParametersData.Filters.Page = a.getSingleIntegerParameter(queryParameters, "page", 1, v)
	queryParametersData.Filters.PageSize = a.getSingleIntegerParameter(queryParameters, "page_size", 10, v)

	queryParametersData.Filters.Sort = a.getSingleQueryParameter(queryParameters, "sort", "id")
	queryParametersData.Filters.SortSafeList = []string{"id", "name", "address", "phone", "email", "-id", "-name", "-address", "-phone", "-email"}

	
	data.ValidateFilters(v, queryParametersData.Filters)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	trainers, metadata, err := a.trainerModel.GetAll(queryParametersData.User_id, queryParametersData.Name, queryParametersData.Address, queryParametersData.Phone, queryParametersData.Email, queryParametersData.Filters)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{
		"trainers":   trainers,
		"@metadata": metadata,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
