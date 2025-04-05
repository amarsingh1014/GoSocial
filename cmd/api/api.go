package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"social/internal/auth"
	"social/internal/mailer"
	"social/internal/store"
	"social/internal/store/cache"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"go.uber.org/zap"

	"social/docs"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type application struct {
	config        config
	store         *store.Storage
	cacheStorage  cache.Storage
	logger        *zap.SugaredLogger
	mailer        mailer.Client
	authenticator auth.Authenticator
}

type config struct {
	addr        string
	dbconn      dbconfig
	env         string
	apiURL      string
	mail        mailconfig
	frontendURL string
	auth        authconfig
	redisCfg    redisConfig
}

type redisConfig struct {
	addr    string
	pw      string
	db      int
	enabled bool
}

type authconfig struct {
	basic basicconfig
	token tokenconfig
}

type tokenconfig struct {
	secret string
	aud    string
	iss    string
	exp    time.Duration
}

type basicconfig struct {
	user     string
	password string
}

type mailconfig struct {
	sendGrid  sendGridConfig
	fromEmail string
	exp       time.Duration
}

type sendGridConfig struct {
	apiKey string
}

type dbconfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdletime  string
}

func (app *application) mount() *chi.Mux {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	r.Route("/v1", func(r chi.Router) {
		// Public routes
		r.Get("/health", app.healthCheckHandler)
		docsURL := fmt.Sprintf("%s/swagger/doc.json", app.config.apiURL)
		r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(docsURL)))
		r.Route("/authentication", func(r chi.Router) {
			r.Post("/user", app.registerUserHandler)
			r.Post("/token", app.createTokenHandler)
		})
		r.Put("/users/activate/{token}", app.activateUserHandler)

		// All authenticated routes
		r.Group(func(r chi.Router) {
			r.Use(app.AuthTokenMiddleware) // Auth middleware applied once here

			// Posts routes
			r.Route("/posts", func(r chi.Router) {
				r.Post("/", app.createPostsHandler)

				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", app.getPostHandler)
					r.Delete("/", app.checkPostOwnership("admin", app.deletePostHandler))
					r.Patch("/", app.checkPostOwnership("moderator", app.updatePostHandler))
					r.Post("/comment", app.checkPostOwnership("user", app.createCommentHandler))
				})
			})

			// User routes
			r.Route("/users", func(r chi.Router) {
				r.Get("/feed", app.getUserFeedHandler)

				r.Route("/{userID}", func(r chi.Router) {
					r.Get("/", app.getUserHandler)
					r.Put("/follow", app.followUserHandler)
					r.Put("/unfollow", app.unfollowUserHandler)
				})
			})
		})
	})

	return r // Add this return statement
}

func (app *application) run(mux *chi.Mux) error {

	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.Host = app.config.apiURL
	docs.SwaggerInfo.BasePath = "/v1"

	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	shutdown := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)

		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		app.logger.Infow("signal caught", "signal", s.String())

		shutdown <- srv.Shutdown(ctx)
	}()

	app.logger.Infow("Starting server",
		"version", version,
		"addr", app.config.addr,
		"env", app.config.env,
	)

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdown
	if err != nil {
		app.logger.Errorw("server shutdown failed", "error", err)
		return err
	}
	app.logger.Infow("server shutdown complete")
	return nil
}
