package main

import (
	"errors"
	"log"
	"net/http"
	"social/internal/store"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type CreatePostPayload struct {
	Title   string   `json:"title" validate:"required,max=100"`
	Content string   `json:"content" validate:"required,max=1000"`
	Tags    []string `json:"tags"`
}

// CreatePost godoc
//
//	@Summary		Creates a post
//	@Description	Creates a post
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		CreatePostPayload	true	"Post payload"
//	@Success		201		{object}	store.Post
//	@Failure		400		{object}	error
//	@Failure		401		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/posts [post]
func (app *application) createPostsHandler(w http.ResponseWriter, r *http.Request) {

	var payload CreatePostPayload

	if err := readJSON(w, r, &payload); err != nil {
		app.internalServerError(w, r, err.Error())
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestError(w, r, err.Error())
		return
	}

	user, err := getUserFromContext(r.Context())

	if err != nil {
		app.internalServerError(w, r, err.Error())
		return
	}

	post := &store.Post{
		Title:   payload.Title,
		Content: payload.Content,
		//Todo : change after auth
		Tags:   payload.Tags,
		UserId: user.ID,
	}

	ctx := r.Context()

	if err := app.store.Posts.Create(ctx, post); err != nil {
		app.internalServerError(w, r, err.Error())
		return
	}

	if err := writeJSON(w, http.StatusCreated, post); err != nil {
		app.internalServerError(w, r, err.Error())
		return
	}
}

// GetPost godoc
//
//	@Summary		Fetches a post
//	@Description	Fetches a post by ID
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"Post ID"
//	@Success		200	{object}	store.Post
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//	@Security		ApiKeyAuth
//	@Router			/posts/{id} [get]
func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	id := chi.URLParam(r, "id")

	idAsInt, err := strconv.Atoi(id)

	if err != nil {
		app.badRequestError(w, r, err.Error())
		return
	}

	post, err := app.store.Posts.GetById(ctx, idAsInt)

	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundError(w, r, err.Error())
		default:
			app.internalServerError(w, r, err.Error())
		}
		return
	}

	comments, err := app.store.Comments.GetByPostID(ctx, post.ID)

	if err != nil {
		app.internalServerError(w, r, err.Error())
		return
	}

	post.Comments = comments

	if err := writeJSON(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err.Error())
		return
	}
}

// DeletePost godoc
//
//	@Summary		Deletes a post
//	@Description	Delete a post by ID
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"Post ID"
//	@Success		204	{object} string
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//	@Security		ApiKeyAuth
//	@Router			/posts/{id} [delete]
func (app *application) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := chi.URLParam(r, "id")

	idAsInt, err := strconv.Atoi(id)

	if err != nil {
		app.internalServerError(w, r, err.Error())
		return
	}

	post, err := app.store.Posts.GetById(ctx, idAsInt)

	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundError(w, r, err.Error())
		default:
			app.internalServerError(w, r, err.Error())
		}
		return
	}

	if err := app.store.Posts.Delete(ctx, post.ID); err != nil {
		app.internalServerError(w, r, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type updatePostPayload struct {
	Title   *string `json:"title" validate:"omitempty,max=100"`
	Content *string `json:"content" validate:"omitempty,max=1000"`
}

// UpdatePost godoc
//
//	@Summary		Updates a post
//	@Description	Updates a post by ID
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"Post ID"
//	@Param			payload	body		updatePostPayload	true	"Post payload"
//	@Success		200		{object}	store.Post
//	@Failure		400		{object}	error
//	@Failure		401		{object}	error
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/posts/{id} [patch]
func (app *application) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := chi.URLParam(r, "id")

	idAsInt, err := strconv.Atoi(id)

	if err != nil {
		app.badRequestError(w, r, err.Error())
		return
	}

	post, err := app.store.Posts.GetById(ctx, idAsInt)

	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundError(w, r, err.Error())
		default:
			app.internalServerError(w, r, err.Error())
		}
		return
	}

	var payload updatePostPayload

	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestError(w, r, err.Error())
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestError(w, r, err.Error())
		return
	}

	log.Printf("Payload: %+v", payload)

	if payload.Content != nil {
		post.Content = *payload.Content
	}

	if payload.Title != nil {
		post.Title = *payload.Title
	}

	if err := app.store.Posts.Update(ctx, post); err != nil {
		app.internalServerError(w, r, err.Error())
		return
	}

	if err := writeJSON(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err.Error())
		return
	}

}

type createCommentPayload struct {
	Content string `json:"content" validate:"required,max=1000"`
	UserID  int    `json:"user_id" validate:"required"`
}

func (app *application) createCommentHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	id := chi.URLParam(r, "id")

	idAsInt, err := strconv.Atoi(id)

	if err != nil {
		app.badRequestError(w, r, err.Error())
		return
	}

	var payload createCommentPayload

	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestError(w, r, err.Error())
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestError(w, r, err.Error())
		return
	}

	comment := &store.Comment{
		Content: payload.Content,
		PostID:  idAsInt,
		UserID:  payload.UserID,
	}

	if err := app.store.Comments.Create(ctx, comment); err != nil {
		app.internalServerError(w, r, err.Error())
		return
	}

	if err := writeJSON(w, http.StatusCreated, comment); err != nil {
		app.internalServerError(w, r, err.Error())
		return
	}

}

// TODO : Add the delete comment handler method
// TODO : Add the middleware to fetch the user from the context
