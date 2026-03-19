package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Joseph-Koop/json-project/internal/data"
	"github.com/Joseph-Koop/json-project/internal/validator"
)

func (a *applicationDependencies) postSessionTimeHandler(w http.ResponseWriter, r *http.Request) {
	var incomingData struct {
		Class_id            int `json:"class_id"`
		Day             string `json:"day"`
		Time            	time.Time `json:"time"`
		Duration            int `json:"duration"`
	}
	
	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}
	
	sessionTime := &data.SessionTime{
		Class_id: 		incomingData.Class_id,
		Day: 	incomingData.Day,
		Time: 		incomingData.Time,
		Duration: 		incomingData.Duration,
	}
	
	v := validator.New()
	
	data.ValidateSessionTime(v, sessionTime)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors) 
		return
	}
	
	err = a.sessionTimeModel.Insert(sessionTime)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/sessionTimes/%d", sessionTime.ID))

	data := envelope{
		"sessionTime": sessionTime,
	}
	err = a.writeJSON(w, http.StatusCreated, data, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}

func (a *applicationDependencies) displaySessionTimeHandler(w http.ResponseWriter, r *http.Request) {
	
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	sessionTime, err := a.sessionTimeModel.Get(id)
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
		"sessionTime": sessionTime,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}

func (a *applicationDependencies) updateSessionTimeHandler(w http.ResponseWriter, r *http.Request) {
	
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	sessionTime, err := a.sessionTimeModel.Get(id)
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
		Day     *string `json:"day"`
		Time      	*time.Time `json:"time"`
		Duration      	*int `json:"duration"`
	}

	
	err = a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}
	
	if incomingData.Class_id != nil {
		sessionTime.Class_id = *incomingData.Class_id
	}
	
	if incomingData.Day != nil {
		sessionTime.Day = *incomingData.Day
	}
	
	if incomingData.Time != nil {
		sessionTime.Time = *incomingData.Time
	}
	
	if incomingData.Duration != nil {
		sessionTime.Duration = *incomingData.Duration
	}
	
	v := validator.New()
	data.ValidateSessionTime(v, sessionTime)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	
	err = a.sessionTimeModel.Update(sessionTime)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	data := envelope{
		"sessionTime": sessionTime,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}

func (a *applicationDependencies) deleteSessionTimeHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	err = a.sessionTimeModel.Delete(id)

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
		"message": "SessionTime successfully deleted.",
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}

}

func (a *applicationDependencies) listSessionTimesHandler(w http.ResponseWriter, r *http.Request) {
	
	var queryParametersData struct {
		Class_id     	*int
		Day   			*string
		Time     		*time.Time
		Duration     	*int
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

	day_string := a.getSingleQueryParameter(queryParameters, "day", "")
	if day_string != "" {
		queryParametersData.Day = &day_string
	}

	time_string := a.getSingleQueryParameter(queryParameters, "time", "")
	if time_string != "" {
		parsed_time, err := time.Parse("15:04", time_string)
		if err != nil {
			v.AddError(time_string, "Must be an valid time.")
		} else {
			queryParametersData.Time = &parsed_time
		}
	}

	duration_string := a.getSingleQueryParameter(queryParameters, "phone", "")
	if duration_string != "" {
		duration_int, err := strconv.Atoi(duration_string)

		if err == nil && duration_int != 0 {
			queryParametersData.Duration = &duration_int
		} else {
			v.AddError(duration_string, "Must be an integer value.")
		}
	}

	queryParametersData.Filters.Page = a.getSingleIntegerParameter(queryParameters, "page", 1, v)
	queryParametersData.Filters.PageSize = a.getSingleIntegerParameter(queryParameters, "page_size", 10, v)

	queryParametersData.Filters.Sort = a.getSingleQueryParameter(queryParameters, "sort", "id")
	queryParametersData.Filters.SortSafeList = []string{"id", "class_id", "day", "time", "duration", "-id", "-class_id", "-day", "-time", "-duration"}

	
	data.ValidateFilters(v, queryParametersData.Filters)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	sessionTimes, metadata, err := a.sessionTimeModel.GetAll(queryParametersData.Class_id, queryParametersData.Day, queryParametersData.Time, queryParametersData.Duration, queryParametersData.Filters)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{
		"sessionTimes":   sessionTimes,
		"@metadata": metadata,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
