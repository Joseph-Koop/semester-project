package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Joseph-Koop/json-project/internal/data"
	"github.com/Joseph-Koop/json-project/internal/validator"
)

func (a *applicationDependencies) postAttendanceHandler(w http.ResponseWriter, r *http.Request) {
	var incomingData struct {
		Registration_id             int `json:"registration_id"`
		Session_id          int `json:"session_id"`
	}
	
	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}
	
	attendance := &data.Attendance{
		Registration_id: 		incomingData.Registration_id,
		Session_id: 		incomingData.Session_id,
	}
	
	v := validator.New()
	
	a.attendanceModel.ValidateAttendance(v, attendance)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors) 
		return
	}
	
	err = a.attendanceModel.Insert(attendance)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/attendances/%d", attendance.ID))

	data := envelope{
		"attendance": attendance,
	}
	err = a.writeJSON(w, http.StatusCreated, data, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}

func (a *applicationDependencies) displayAttendanceHandler(w http.ResponseWriter, r *http.Request) {
	
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	attendance, err := a.attendanceModel.Get(id)
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
		"attendance": attendance,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}

func (a *applicationDependencies) updateAttendanceHandler(w http.ResponseWriter, r *http.Request) {
	
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	attendance, err := a.attendanceModel.Get(id)
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
		Registration_id      	*int `json:"registration_id"`
		Session_id      	*int `json:"session_id"`
	}

	
	err = a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}
	
	if incomingData.Registration_id != nil {
		attendance.Registration_id = *incomingData.Registration_id
	}
	
	if incomingData.Session_id != nil {
		attendance.Session_id = *incomingData.Session_id
	}
	
	v := validator.New()
	a.attendanceModel.ValidateAttendance(v, attendance)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	
	err = a.attendanceModel.Update(attendance)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	data := envelope{
		"attendance": attendance,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}

func (a *applicationDependencies) deleteAttendanceHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	err = a.attendanceModel.Delete(id)

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
		"message": "Attendance successfully deleted.",
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}

}

func (a *applicationDependencies) listAttendancesHandler(w http.ResponseWriter, r *http.Request) {
	
	var queryParametersData struct {
		Registration_id     	*int
		Session_id     	*int
		data.Filters
	}
	
	queryParameters := r.URL.Query()

	v := validator.New()

	registration_id_string := a.getSingleQueryParameter(queryParameters, "registration_id", "")
	if registration_id_string != "" {
		registration_id_int, err := strconv.Atoi(registration_id_string)

		if err == nil && registration_id_int != 0 {
			queryParametersData.Registration_id = &registration_id_int
		} else {
			v.AddError(registration_id_string, "Must be an integer value.")
		}
	}

	session_id_string := a.getSingleQueryParameter(queryParameters, "session_id", "")
	if session_id_string != "" {
		session_id_int, err := strconv.Atoi(session_id_string)

		if err == nil && session_id_int != 0 {
			queryParametersData.Session_id = &session_id_int
		} else {
			v.AddError(session_id_string, "Must be an integer value.")
		}
	}

	queryParametersData.Filters.Page = a.getSingleIntegerParameter(queryParameters, "page", 1, v)
	queryParametersData.Filters.PageSize = a.getSingleIntegerParameter(queryParameters, "page_size", 10, v)

	queryParametersData.Filters.Sort = a.getSingleQueryParameter(queryParameters, "sort", "id")
	queryParametersData.Filters.SortSafeList = []string{"id", "registration_id", "session_id", "-id", "-registration_id", "-session_id"}

	
	data.ValidateFilters(v, queryParametersData.Filters)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	attendances, metadata, err := a.attendanceModel.GetAll(queryParametersData.Registration_id, queryParametersData.Session_id, queryParametersData.Filters)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{
		"attendances":   attendances,
		"@metadata": metadata,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
