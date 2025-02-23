package main

import (
	"log"
	"net/http"
	"social/internal/store"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type application struct {
	config config
	store *store.Storage
}

type config struct {
	addr string
	dbconn dbconfig
	env string
}

type dbconfig struct {
	addr string
	maxOpenConns int
	maxIdleConns int
	maxIdletime string
}

func (app *application) mount() *chi.Mux {
	r:= chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", app.healthCheckHandler)

		r.Route("/posts", func(r chi.Router) {
			r.Post("/", app.createPostsHandler)

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", app.getPostHandler)

				r.Delete("/", app.deletePostHandler)
				r.Patch("/", app.updatePostHandler)
				r.Post("/comment", app.createCommentHandler)
			})
		})

		r.Route("/users", func(r chi.Router) {
			// r.Get("/", app.getUsersHandler)
			
			r.Route("/{userID}", func(r chi.Router) {
				r.Use(app.userContextMiddleware)

				r.Get("/", app.getUserHandler)
				r.Put("/follow", app.followUserHandler)
				r.Put("/unfollow", app.unfollowUserHandler)
			})
			
			r.Group(func(r chi.Router) {
				r.Get("/feed", app.getUserFeedHandler)
			})
		})
	})

	return r
}

func (app *application) run(mux *chi.Mux) error {

	srv := &http.Server{
		Addr:    app.config.addr,
		Handler : mux,
		WriteTimeout: 15 * time.Second,
		ReadTimeout: 10 * time.Second,	
		IdleTimeout: 120 * time.Second,
	}

	log.Printf("Server is running on %s", app.config.addr)

	return srv.ListenAndServe()
}