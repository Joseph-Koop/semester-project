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
	router.HandlerFunc(http.MethodGet, "/classes", a.viewClassHandler)
	router.HandlerFunc(http.MethodPost, "/classes/post", a.postClassHandler)
	// router.HandlerFunc(http.MethodPut, "/classes/:id/put", a.putClassHandler)
	// router.HandlerFunc(http.MethodPatch, "/classes/:id/patch", a.patchClassHandler)
	// router.HandlerFunc(http.MethodDelete, "/classes/:id/delete", a.deletelassHandler)

	return a.recoverPanic(router)

}
