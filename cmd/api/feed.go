package main

import (
	"net/http"
	"social/internal/store"
)

func(app *application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {
	
	fq := store.PaginatedFieldQuery{
		Limit: 5,
		Offset: 1,
		Sort: "desc",
	}

	fq, err := fq.Parse(r)

	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Bad Request")
		return
	}

	if err := Validate.Struct(fq); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Bad Request")
		return
	}

	ctx := r.Context()

	feed, err := app.store.Posts.GetUserFeed(ctx, int64(188), fq)

	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	if err := writeJSON(w, http.StatusOK, feed); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal Server Error")
	}

}