package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Joseph-Koop/json-project/internal/data"
	"github.com/Joseph-Koop/json-project/internal/validator"
)

func (a *applicationDependencies) postSessionHandler(w http.ResponseWriter, r *http.Request) {
	var incomingData struct {
		Class_id            int `json:"class_id"`
	}
	
	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}
	
	session := &data.Session{
		Class_id: 		incomingData.Class_id,
	}
	
	v := validator.New()
	
	data.ValidateSession(v, session)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors) 
		return
	}
	
	err = a.sessionModel.Insert(session)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/sessions/%d", session.ID))

	data := envelope{
		"session": session,
	}
	err = a.writeJSON(w, http.StatusCreated, data, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}

func (a *applicationDependencies) displaySessionHandler(w http.ResponseWriter, r *http.Request) {
	
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	session, err := a.sessionModel.Get(id)
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
		"session": session,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}

func (a *applicationDependencies) updateSessionHandler(w http.ResponseWriter, r *http.Request) {
	
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	session, err := a.sessionModel.Get(id)
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
		Class_id      	*int `json:"class_id"`
	}

	
	err = a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}
	
	if incomingData.Class_id != nil {
		session.Class_id = *incomingData.Class_id
	}
	
	v := validator.New()
	data.ValidateSession(v, session)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	
	err = a.sessionModel.Update(session)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	data := envelope{
		"session": session,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}

func (a *applicationDependencies) deleteSessionHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	err = a.sessionModel.Delete(id)

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
		"message": "Session successfully deleted.",
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}

}

func (a *applicationDependencies) listSessionsHandler(w http.ResponseWriter, r *http.Request) {
	
	var queryParametersData struct {
		Class_id     	*int
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

	queryParametersData.Filters.Page = a.getSingleIntegerParameter(queryParameters, "page", 1, v)
	queryParametersData.Filters.PageSize = a.getSingleIntegerParameter(queryParameters, "page_size", 10, v)

	queryParametersData.Filters.Sort = a.getSingleQueryParameter(queryParameters, "sort", "id")
	queryParametersData.Filters.SortSafeList = []string{"id", "class_id", "-id", "-class_id"}

	
	data.ValidateFilters(v, queryParametersData.Filters)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	user := a.contextGetUser(r)

	var sessions []*data.Session
	var err error
	var metadata any

	switch user.Role_id{
		case 1:
			// Admin → all sessions
			sessions, metadata, err = a.sessionModel.GetAll(queryParametersData.Class_id, queryParametersData.Filters)


		case 2:
			// Trainer → sessions of their classes
			trainer, err2 := a.trainerModel.GetByUserID(user.ID)
			if err2 != nil {
				a.serverErrorResponse(w, r, err2)
				return
			}

			sessions, metadata, err = a.sessionModel.GetAllByTrainerID(trainer.ID, queryParametersData.Class_id, queryParametersData.Filters)

		case 3:
			// Member → sessions they are registered/attending
			member, err2 := a.memberModel.GetByUserID(user.ID)
			if err2 != nil {
				a.serverErrorResponse(w, r, err2)
				return
			}

			sessions, metadata, err = a.sessionModel.GetAllByMemberID(member.ID, queryParametersData.Class_id, queryParametersData.Filters)
		default:
			a.notPermittedResponse(w, r)
			return
	}

	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{
		"sessions":   sessions,
		"@metadata": metadata,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
