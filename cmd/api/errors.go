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

func (a *applicationDependencies) errorResponseJSON(w http.ResponseWriter, r *http.Request, status int, message any) {

	errorData := envelope{"error": message}
	err := a.writeJSON(w, status, errorData, nil)
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
	a.errorResponseJSON(w, r, http.StatusInternalServerError, message)
}

// send an error response if our client messes up with a 404
func (a *applicationDependencies) notFoundResponse(w http.ResponseWriter, r *http.Request) {

	// we only log server errors, not client errors
	// prepare a response to send to the client
	message := "The requested resource could not be found."
	a.errorResponseJSON(w, r, http.StatusNotFound, message)
}

// send an error response if our client messes up with a 405
func (a *applicationDependencies) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {

	// we only log server errors, not client errors
	// prepare a formatted response to send to the client
	message := fmt.Sprintf("The %s method is not supported for this resource.", r.Method)

	a.errorResponseJSON(w, r, http.StatusMethodNotAllowed, message)
}

// send an error response if our client messes up with a 400 (bad request)
func (a *applicationDependencies) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {

	a.errorResponseJSON(w, r, http.StatusBadRequest, err.Error())
}

func (a *applicationDependencies)failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
     a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, errors)
}


// func (a *applicationDependencies) viewClassHandler(w http.ResponseWriter, r *http.Request) {
// 	// panic("Panic!!!!")   // deliberate panic
// 	data := envelope{
// 		"classes": []map[string]any{
// 			{
// 				"id":              1,
// 				"studio":          "Pool",
// 				"trainer":         "Mo Lester",
// 				"capacity_limit":  10,
// 				"membership_tier": "premium",
// 				"name":            "Swimming Classes",
// 				"status":          "Ongoing",
// 			},
// 			{
// 				"id":              2,
// 				"studio":          "Studio #3",
// 				"trainer":         "Ben Dover",
// 				"capacity_limit":  20,
// 				"membership_tier": "basic",
// 				"name":            "Powerlifting Classes",
// 				"status":          "Ongoing",
// 			},
// 			{
// 				"id":              3,
// 				"studio":          "Studio #6",
// 				"trainer":         "Anita Bath",
// 				"capacity_limit":  20,
// 				"membership_tier": "basic",
// 				"name":            "Yoga Classes",
// 				"status":          "Paused",
// 			},
// 		},
// 	}
// 	err := a.writeJSON(w, http.StatusOK, data, nil)
// 	if err != nil {
// 		a.serverErrorResponse(w, r, err)
// 	}
// }