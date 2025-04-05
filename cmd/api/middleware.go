package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"social/internal/store"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
)

func (app *application) AuthTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//parse the token
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			app.unauthorizedError(w, r, "no authorization header provided")
			return
		}

		parts := strings.Split(authHeader, " ")

		if len(parts) != 2 || parts[0] != "Bearer" {
			app.unauthorizedError(w, r, "malformed header")
			return
		}

		//validate the token
		token := parts[1]

		jwtToken, err := app.authenticator.ValidateToken(token)
		if err != nil {
			app.unauthorizedError(w, r, "error validating token")
			return
		}

		claims := jwtToken.Claims.(jwt.MapClaims)

		userID, err := strconv.ParseInt(fmt.Sprintf("%.f", claims["sub"]), 10, 64)
		if err != nil {
			app.unauthorizedError(w, r, "error parsing user id")
			return
		}

		ctx := r.Context()

		user, err := app.getUser(ctx, int(userID))
		if err != nil {
			app.internalServerError(w, r, err.Error())
			return
		}

		if user == nil {
			app.unauthorizedError(w, r, "user not found in system")
			return
		}

		ctx = context.WithValue(ctx, userContext, user)
		next.ServeHTTP(w, r.WithContext(ctx))
		//if the token is invalid, return an unauthorized error
	})
}

func (app *application) BasicAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				app.unauthorizedBasicError(w, r, "no authorization header provided")
				return
			}

			//parse the header
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Basic" {
				app.unauthorizedBasicError(w, r, "invalid authorization header")
				return
			}

			// decode the base64 encoded string
			decoded, err := base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				app.unauthorizedBasicError(w, r, "error decoding authorization header")
				return
			}

			// check if the username and password are correct
			creds := strings.SplitN(string(decoded), ":", 2)

			username := app.config.auth.basic.user
			password := app.config.auth.basic.password

			if len(creds) != 2 || creds[0] != username || creds[1] != password {
				app.unauthorizedBasicError(w, r, "invalid username or password")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (app *application) checkPostOwnership(requiredRole string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check if post belongs to the user
		user := r.Context().Value(userContext).(*store.User)
		postID, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			app.badRequestError(w, r, err.Error())
			return
		}

		post, err := app.store.Posts.GetById(r.Context(), postID)

		if err != nil {
			app.notFoundError(w, r, "post don't exist")
			return
		}

		if post.UserId == user.ID {
			next.ServeHTTP(w, r)
			return
		}

		//role precedance
		allowed, err := app.checkRolePrecedance(r.Context(), user, requiredRole)

		if err != nil {
			app.internalServerError(w, r, err.Error())
			return
		}

		if !allowed {
			app.forbiddenError(w, r, "forbidden")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) checkRolePrecedance(ctx context.Context, user *store.User, rolename string) (bool, error) {
	role, err := app.store.Roles.GetByName(ctx, rolename)

	if err != nil {
		return false, err
	}

	return user.Role.Level >= role.Level, nil
}

func (app *application) getUser(ctx context.Context, userID int) (*store.User, error) {
	// Don't log "cache hit" until we actually confirm it's a hit
	user, err := app.cacheStorage.Users.Get(ctx, int64(userID))

	// Redis miss or error - both should trigger DB lookup
	if err == redis.Nil || user == nil {
		app.logger.Infow("cache miss", "key", "user", "id", userID)
		user, err = app.store.Users.GetById(ctx, userID)
		if err != nil {
			return nil, err
		}

		// If still nil after DB lookup, return an error
		if user == nil {
			return nil, fmt.Errorf("user with ID %d not found in database", userID)
		}

		// Cache the user for future requests
		if err = app.cacheStorage.Users.Set(ctx, user); err != nil {
			app.logger.Warnw("Failed to cache user", "error", err)
			// Continue anyway, just couldn't cache
		}
	} else if err != nil {
		return nil, err
	} else {
		app.logger.Infow("cache hit", "key", "user", "id", userID)
	}

	return user, nil
}
