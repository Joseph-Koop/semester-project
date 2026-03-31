package main

import (
	"fmt"
	"net/http"
)

func (a *applicationDependencies) logError(r *http.Request, err error) {

	method := r.Method
	uri := r.URL.RequestURI()
	a.logger.Error(err.Error(), "method", method, "uri", uri)

}

func (a *applicationDependencies) errorResponseJSON(w http.ResponseWriter, r *http.Request, status int, message any, headers http.Header) {

	errorData := envelope{"error": message}
	err := a.writeJSON(w, status, errorData, headers)
	if err != nil {
		a.logError(r, err)
		w.WriteHeader(500)
	}
}

func (a *applicationDependencies) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {

	// first thing is to log error message
	a.logError(r, err)
	// prepare a response to send to the client
	message := "The server encountered a problem and could not process your request."
	a.errorResponseJSON(w, r, http.StatusInternalServerError, message, nil)
}

// send an error response if our client messes up with a 404
func (a *applicationDependencies) notFoundResponse(w http.ResponseWriter, r *http.Request) {

	// we only log server errors, not client errors
	// prepare a response to send to the client
	message := "The requested resource could not be found."
	a.errorResponseJSON(w, r, http.StatusNotFound, message, nil)
}

// send an error response if our client messes up with a 405
func (a *applicationDependencies) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {

	// we only log server errors, not client errors
	// prepare a formatted response to send to the client
	message := fmt.Sprintf("The %s method is not supported for this resource.", r.Method)

	a.errorResponseJSON(w, r, http.StatusMethodNotAllowed, message, nil)
}

// send an error response if our client messes up with a 400 (bad request)
func (a *applicationDependencies) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {

	a.errorResponseJSON(w, r, http.StatusBadRequest, err.Error(), nil)
}

func (a *applicationDependencies)failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, errors, nil)
}

// send an error response if rate limit exceeded (429 - Too Many Requests)
func (a *applicationDependencies)rateLimitExceededResponse(w http.ResponseWriter, r *http.Request)  {
    headers := make(http.Header)
    headers.Set("Retry-After", "5") 
    message := "Rate limit exceeded."
    a.errorResponseJSON(w, r, http.StatusTooManyRequests, message, headers)
}

// send an error response if we have an edit conflict status 409 
func (a *applicationDependencies)editConflictResponse(w http.ResponseWriter, r *http.Request)  {

    message := "Unable to update the record due to an edit conflict, please try again."
a.errorResponseJSON(w, r, http.StatusConflict, message, nil)

}

