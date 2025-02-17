package main

import (
	"errors"
	"net/http"
	"social/internal/store"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type CreatePostPayload struct {
	Title string `json:"title" validate:"required,max=100"`
	Content string `json:"content" validate:"required,max=1000"`
	Tags []string `json:"tags"`
}

func (app *application) createPostsHandler(w http.ResponseWriter, r *http.Request) {

	var payload CreatePostPayload

	if err := readJSON(w, r, &payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Bad Request")
		return
	}

	if err := Validate.Struct(payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	post := &store.Post{
		Title: payload.Title,
		Content: payload.Content,
		//Todo : change after auth 
		Tags: payload.Tags,
		UserId: 2,
	}

	ctx := r.Context()

	if err := app.store.Posts.Create(ctx, post); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	if err := writeJSON(w, http.StatusCreated, post); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
}

func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	id := chi.URLParam(r, "id")

	idAsInt, err := strconv.Atoi(id)

	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Bad Request")
		return
	}

	post, err := app.store.Posts.GetById(ctx, idAsInt)

	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			writeJSONError(w, http.StatusNotFound, err.Error())
		default:
			writeJSONError(w, http.StatusInternalServerError, "Internal Server Error")
		}
		return 
	}

	comments, err := app.store.Comments.GetByPostID(ctx, post.ID)

	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "here")
		return
	}

	post.Comments = comments

	if err := writeJSON(w, http.StatusOK, post); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
}

func (app *application) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := chi.URLParam(r, "id")

	idAsInt, err := strconv.Atoi(id)

	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Bad Request")
		return
	}

	post, err := app.store.Posts.GetById(ctx, idAsInt)

	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			writeJSONError(w, http.StatusNotFound, err.Error())
		default:
			writeJSONError(w, http.StatusInternalServerError, "Internal Server Error")
		}
		return
	}

	if err := app.store.Posts.Delete(ctx, post.ID); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}


// TODO : update the post handler method
func (app *application) updatePostHandler(w http.ResponseWriter,r *http.Request) {
	ctx := r.Context()

	id := chi.URLParam(r, "id")

	idAsInt, err := strconv.Atoi(id)

	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Bad Request")
		return
	}

	post, err := app.store.Posts.GetById(ctx, idAsInt)

	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			writeJSONError(w, http.StatusNotFound, err.Error())
		default:
			writeJSONError(w, http.StatusInternalServerError, "Internal Server Error")
		}
		return
	}

	var payload CreatePostPayload

	if err := readJSON(w, r, &payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Bad Request")
		return
	}

	if err := Validate.Struct(payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	post.Content = payload.Content

	post.Title = payload.Title

	post.Tags = payload.Tags

	if err := app.store.Posts.Update(ctx, post); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	if err := writeJSON(w, http.StatusOK, post); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

}