package main

import (
	"net/http"
	"expvar"

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

	router.HandlerFunc(http.MethodGet, "/gyms", a.listGymsHandler)
	router.HandlerFunc(http.MethodGet, "/gyms/:id", a.displayGymHandler)
	router.HandlerFunc(http.MethodPost, "/gyms/add", a.postGymHandler)
	router.HandlerFunc(http.MethodPatch, "/gyms/:id/update", a.updateGymHandler)
	router.HandlerFunc(http.MethodDelete, "/gyms/:id/delete", a.deleteGymHandler)

	router.HandlerFunc(http.MethodGet, "/trainers", a.listTrainersHandler)
	router.HandlerFunc(http.MethodGet, "/trainers/:id", a.displayTrainerHandler)
	router.HandlerFunc(http.MethodPost, "/trainers/add", a.postTrainerHandler)
	router.HandlerFunc(http.MethodPatch, "/trainers/:id/update", a.updateTrainerHandler)
	router.HandlerFunc(http.MethodDelete, "/trainers/:id/delete", a.deleteTrainerHandler)

	router.HandlerFunc(http.MethodGet, "/members", a.listMembersHandler)
	router.HandlerFunc(http.MethodGet, "/members/:id", a.displayMemberHandler)
	router.HandlerFunc(http.MethodPost, "/members/add", a.postMemberHandler)
	router.HandlerFunc(http.MethodPatch, "/members/:id/update", a.updateMemberHandler)
	router.HandlerFunc(http.MethodDelete, "/members/:id/delete", a.deleteMemberHandler)

	router.HandlerFunc(http.MethodGet, "/studios", a.listStudiosHandler)
	router.HandlerFunc(http.MethodGet, "/studios/:id", a.displayStudioHandler)
	router.HandlerFunc(http.MethodPost, "/studios/add", a.postStudioHandler)
	router.HandlerFunc(http.MethodPatch, "/studios/:id/update", a.updateStudioHandler)
	router.HandlerFunc(http.MethodDelete, "/studios/:id/delete", a.deleteStudioHandler)

	router.Handler(http.MethodGet, "/metrics", expvar.Handler())

	
	// router.HandlerFunc(http.MethodPut, "/classes/:id/put", a.updateClassHandler)

	return a.logRequest(a.metrics(a.recoverPanic(a.compressResponse(a.enableCORS(a.rateLimit(router))))))

}
