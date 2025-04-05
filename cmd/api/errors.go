package main

import (
	"net/http"
)

func(app *application) badRequestError(w http.ResponseWriter, r *http.Request, error string) {
	
	app.logger.Errorw("bad request", "method", r.Method, r.URL.Path, "error", error)
	writeJSONError(w, http.StatusBadRequest, error)
}

func(app *application) internalServerError(w http.ResponseWriter, r *http.Request, error string) {
	
	app.logger.Warnf("internal error", "method", r.Method, r.URL.Path, "error", error)
	writeJSONError(w, http.StatusInternalServerError, error)

}

func(app *application) notFoundError(w http.ResponseWriter, r *http.Request, error string) {
	
	app.logger.Warnf("resource not found", "method", r.Method, r.URL.Path, "error", error)
	writeJSONError(w, http.StatusNotFound, error)
}

func (app *application) unauthorizedError(w http.ResponseWriter, r *http.Request, error string) {
	
	app.logger.Warnf("unauthorized", "method", r.Method, r.URL.Path, "error", error)
	writeJSONError(w, http.StatusUnauthorized, error)
}

func (app *application) unauthorizedBasicError(w http.ResponseWriter, r *http.Request, error string) {
	
	app.logger.Warnf("unauthorized", "method", r.Method, r.URL.Path)
	w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

	writeJSONError(w, http.StatusUnauthorized, error)
}

func (app *application) forbiddenError(w http.ResponseWriter, r *http.Request, error string) {
	app.logger.Warnf("forbidden", "method", r.Method, r.URL.Path, "error", error)
	
	writeJSONError(w, http.StatusForbidden, error)
}