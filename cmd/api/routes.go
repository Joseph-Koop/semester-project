package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (a *applicationDependencies) routes() http.Handler {

	// setup a new router
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(a.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(a.methodNotAllowedResponse)

	// setup routes
	router.HandlerFunc(http.MethodGet, "/classes", a.listClassesHandler)
	router.HandlerFunc(http.MethodGet, "/classes/:id", a.displayClassHandler)
	router.HandlerFunc(http.MethodPost, "/classes/add", a.postClassHandler)
	router.HandlerFunc(http.MethodPatch, "/classes/:id/update", a.updateClassHandler)
	router.HandlerFunc(http.MethodDelete, "/classes/:id/delete", a.deleteClassHandler)
	
	// router.HandlerFunc(http.MethodPut, "/classes/:id/put", a.updateClassHandler)

	return a.recoverPanic(a.rateLimit(router))

}
