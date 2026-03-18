package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Joseph-Koop/json-project/internal/data"
	"github.com/Joseph-Koop/json-project/internal/validator"
)

func (a *applicationDependencies) postStudioHandler(w http.ResponseWriter, r *http.Request) {
	var incomingData struct {
		Gym_id             int `json:"gym_id"`
		Name          string `json:"name"`
		Access            string `json:"access"`
	}
	
	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}
	
	studio := &data.Studio{
		Gym_id: 		incomingData.Gym_id,
		Name: 		incomingData.Name,
		Access: 	incomingData.Access,
	}
	
	v := validator.New()
	
	data.ValidateStudio(v, studio)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors) 
		return
	}
	
	err = a.studioModel.Insert(studio)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/studios/%d", studio.ID))

	data := envelope{
		"studio": studio,
	}
	err = a.writeJSON(w, http.StatusCreated, data, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}

func (a *applicationDependencies) displayStudioHandler(w http.ResponseWriter, r *http.Request) {
	
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	studio, err := a.studioModel.Get(id)
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
		"studio": studio,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}

func (a *applicationDependencies) updateStudioHandler(w http.ResponseWriter, r *http.Request) {
	
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	studio, err := a.studioModel.Get(id)
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
		Gym_id      	*int `json:"gym_id"`
		Name      	*string `json:"name"`
		Access     *string `json:"access"`
	}

	
	err = a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}
	
	if incomingData.Gym_id != nil {
		studio.Gym_id = *incomingData.Gym_id
	}
	
	if incomingData.Name != nil {
		studio.Name = *incomingData.Name
	}
	
	if incomingData.Access != nil {
		studio.Access = *incomingData.Access
	}
	
	v := validator.New()
	data.ValidateStudio(v, studio)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	
	err = a.studioModel.Update(studio)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	data := envelope{
		"studio": studio,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}

func (a *applicationDependencies) deleteStudioHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	err = a.studioModel.Delete(id)

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
		"message": "Studio successfully deleted.",
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}

}

func (a *applicationDependencies) listStudiosHandler(w http.ResponseWriter, r *http.Request) {
	
	var queryParametersData struct {
		Gym_id     	*int
		Name     	*string
		Access   	*string
		data.Filters
	}
	
	queryParameters := r.URL.Query()

	v := validator.New()

	gym_id_string := a.getSingleQueryParameter(queryParameters, "gym_id", "")
	if gym_id_string != "" {
		gym_id_int, err := strconv.Atoi(gym_id_string)

		if err == nil && gym_id_int != 0 {
			queryParametersData.Gym_id = &gym_id_int
		} else {
			v.AddError(gym_id_string, "Must be an integer value.")
		}
	}

	name_string := a.getSingleQueryParameter(queryParameters, "name", "")
	if name_string != "" {
		queryParametersData.Name = &name_string
	}

	access_string := a.getSingleQueryParameter(queryParameters, "access", "")
	if access_string != "" {
		queryParametersData.Access = &access_string
	}

	queryParametersData.Filters.Page = a.getSingleIntegerParameter(queryParameters, "page", 1, v)
	queryParametersData.Filters.PageSize = a.getSingleIntegerParameter(queryParameters, "page_size", 10, v)

	queryParametersData.Filters.Sort = a.getSingleQueryParameter(queryParameters, "sort", "id")
	queryParametersData.Filters.SortSafeList = []string{"id", "gym_id", "name", "access", "-id", "-gym_id", "-name", "-access"}

	
	data.ValidateFilters(v, queryParametersData.Filters)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	studios, metadata, err := a.studioModel.GetAll(queryParametersData.Gym_id, queryParametersData.Name, queryParametersData.Access, queryParametersData.Filters)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{
		"studios":   studios,
		"@metadata": metadata,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
