package main

import (
	"context"
	"errors"
	"net/http"
	"social/internal/store"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-redis/redis/v8"
)

type userContextKey string

const userContext userContextKey = "user"

// GetUser godoc
//
//	@Summary		Fetches a user profile
//	@Description	Fetches a user profile by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		int	true	"User ID"
//	@Success		200		{object}	store.User
//	@Failure		400		{object}	error
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/{userID} [get]
func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "userID"))
	if err != nil || userID < 1 {
		app.badRequestError(w, r, "invalid user id")
		return
	}

	ctx := r.Context()

	// First try to get from cache
	user, err := app.getUser(ctx, (userID))

	// If cache miss or error, try database
	if err == redis.Nil || user == nil {
		app.logger.Infow("cache miss", "key", "user", "id", userID)
		user, err = app.store.Users.GetById(ctx, userID)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				app.notFoundError(w, r, err.Error())
				return
			default:
				app.internalServerError(w, r, err.Error())
				return
			}
		}

		// Cache the user for future requests
		if err = app.cacheStorage.Users.Set(ctx, user); err != nil {
			app.logger.Warnw("Failed to cache user", "error", err)
			// Continue anyway, just couldn't cache
		}
	} else if err != nil {
		app.internalServerError(w, r, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, user)
}

type FollowUser struct {
	UserID int64 `json:"user_id"`
}

func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	followedID, err := strconv.Atoi(chi.URLParam(r, "userID"))
	if err != nil {
		app.badRequestError(w, r, err.Error())
		return
	}

	followerUser, err := getUserFromContext(r.Context())
	if err != nil {
		app.internalServerError(w, r, err.Error())
		return
	}
	// Follow the user
	ctx := r.Context()

	if err := app.store.Followers.Follow(ctx, followerUser.ID, int64(followedID)); err != nil {
		switch {
		case errors.Is(err, store.ErrAlreadyFollowing):
			writeJSONError(w, http.StatusConflict, err.Error())
		default:
			app.internalServerError(w, r, err.Error())
		}
		return
	}

	writeJSON(w, http.StatusCreated, nil)

}

func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	unfollowedID, err := strconv.Atoi(chi.URLParam(r, "userID"))
	if err != nil {
		app.badRequestError(w, r, err.Error())
		return
	}

	followerUser, err := getUserFromContext(r.Context())
	if err != nil {
		app.internalServerError(w, r, err.Error())
		return
	}

	ctx := r.Context()

	if err := app.store.Followers.Unfollow(ctx, followerUser.ID, int64(unfollowedID)); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFollowing):
			app.badRequestError(w, r, err.Error())
		default:
			app.internalServerError(w, r, err.Error())
		}
		return
	}

	writeJSON(w, http.StatusNoContent, nil)
}

func (app *application) userContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		id := chi.URLParam(r, "userID")

		idAsInt, err := strconv.Atoi(id)

		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "Bad Request")
			return
		}

		user, err := app.store.Users.GetById(ctx, idAsInt)

		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				writeJSONError(w, http.StatusNotFound, err.Error())
			default:
				writeJSONError(w, http.StatusInternalServerError, "Internal Server Error")
			}
		}

		ctx = context.WithValue(ctx, userContext, user)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUserFromContext(ctx context.Context) (*store.User, error) {
	user, ok := ctx.Value(userContext).(*store.User)
	if !ok {
		return nil, errors.New("user not found in context")
	}
	return user, nil
}

// ActivateUser godoc
//
//	@Summary		Activates a user account
//	@Description	Activates a user account using the activation token
//	@Tags			users
//	@Produce		json
//	@Param			token	path		string	true	"Activation token"
//	@Success		200		{object}	string
//	@Failure		400		{object}	error
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//	@Router			/users/activate/{token} [put]
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	if token == "" {
		app.badRequestError(w, r, "missing token")
		return
	}

	ctx := r.Context()

	if err := app.store.Users.Activate(ctx, token); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundError(w, r, err.Error())
		default:
			app.internalServerError(w, r, err.Error())
		}
		return
	}

	writeJSON(w, http.StatusOK, "user account activated")
}
