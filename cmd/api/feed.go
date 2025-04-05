package main

import (
	"fmt"
	"net/http"
	"social/internal/store"
)

// getUserFeedHandler godoc
//
//	@Summary		Fetches the user feed
//	@Description	Fetches the user feed
//	@Tags			feed
//	@Accept			json
//	@Produce		json
//	@Param			since	query		string	false	"Since"
//	@Param			until	query		string	false	"Until"
//	@Param			limit	query		int		false	"Limit"
//	@Param			offset	query		int		false	"Offset"
//	@Param			sort	query		string	false	"Sort"
//	@Param			tags	query		string	false	"Tags"
//	@Param			search	query		string	false	"Search"
//	@Success		200		{object}	[]store.PostWithMetadata
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/feed [get]
func(app *application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {
	
	fq := store.PaginatedFieldQuery{
		Limit: 5,
		Offset: 0,
		Sort: "desc",
	}

	if err := fq.Parse(r); err != nil {
		app.badRequestError(w, r, err.Error())
		return
	}

	fmt.Printf("Feed query: limit=%d offset=%d search=%s tags=%v\n", 
	fq.Limit, fq.Offset, fq.Search, fq.Tags)	

	if err := Validate.Struct(fq); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Bad Request")
		return
	}

	ctx := r.Context()

	user, err := getUserFromContext(ctx)
	if err != nil {
		app.internalServerError(w, r, err.Error())
		return
	}

	if user == nil {
		app.unauthorizedError(w, r, "user not found")
		return
	}

	fmt.Println("User ID: ", user.ID)

	feed, err := app.store.Posts.GetUserFeed(ctx, int64(user.ID), fq)


	if err != nil {
		app.badRequestError(w, r, err.Error())
		return
	}

	if err := writeJSON(w, http.StatusOK, feed); err != nil {
		app.internalServerError(w, r, err.Error())
		return
	}

}