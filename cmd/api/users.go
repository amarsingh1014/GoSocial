package main

import (
	"context"
	"errors"
	"net/http"
	"social/internal/store"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type userContextKey string
const userContext userContextKey = "user"

func (app *application) getUserHandler(w http.ResponseWriter, r * http.Request) {
	user, err := getUserFromContext(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

type FollowUser struct {
	UserID int64 `json:"user_id"`
}

func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload FollowUser

	if err := readJSON(w, r, &payload); err != nil {
		writeJSONError(w, http.StatusNoContent, "Bad Request")
		return
	}

	followedUser, err := getUserFromContext(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	// Follow the user
	ctx := r.Context()

	if err := app.store.Followers.Follow(ctx, followedUser.ID, payload.UserID); err != nil {
		switch {
		case errors.Is(err, store.ErrAlreadyFollowing):
			writeJSONError(w, http.StatusConflict, err.Error())
		default:
			writeJSONError(w, http.StatusInternalServerError, "Internal Server Error")
		}
		return
	}

	writeJSON(w, http.StatusCreated, nil)

}

func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	unfollowedUser, err := getUserFromContext(r.Context())

	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	var payload FollowUser

	if err := readJSON(w, r, &payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Bad Request")
		return
	}

	ctx := r.Context()

	if err := app.store.Followers.Unfollow(ctx, unfollowedUser.ID, payload.UserID); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFollowing):
			writeJSONError(w, http.StatusBadRequest, err.Error())
		default:
			writeJSONError(w, http.StatusInternalServerError, "Internal Server Error")
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