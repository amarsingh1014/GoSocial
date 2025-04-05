package store

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

type PaginatedFieldQuery struct {
	Limit  int      `json:"limit" validate:"gte=1,lte=100"`
	Offset int      `json:"offset" validate:"gte=0"`
	Sort   string   `json:"sort" validate:"oneof=asc desc"`
	Tags   []string `json:"tags"`
	Search string   `json:"search" validate:"max=100"`
	Since  string   `json:"since"`
	Until  string   `json:"until"`
}

func (p *PaginatedFieldQuery) Parse(r *http.Request) error {

	qs := r.URL.Query()

	limit := qs.Get("limit")

	if limit != "" {
		limit, err := strconv.Atoi(limit)
		if err != nil {
			return err
		}
		p.Limit = limit
	}

	offset := qs.Get("offset")

	if offset != "" {
		offset, err := strconv.Atoi(offset)
		if err != nil {
			return err
		}
		p.Offset = offset
	}

	sort := qs.Get("sort")

	if sort != "" {
		p.Sort = sort
	}

	tags := qs.Get("tags")

	if tags != "" {
		p.Tags = strings.Split(tags, ",")
	}

	search := qs.Get("search")

	if search != "" {
		p.Search = search
	}

	since := qs.Get("since")

	if since != "" {
		p.Since = parseTime(since)
	}

	until := qs.Get("until")

	if until != "" {
		p.Until = parseTime(until)
	}

	return nil
}

func parseTime(s string) string {
	t, err := time.Parse(time.RFC3339, s)

	if err != nil {
		return ""
	}

	return t.Format(time.RFC3339)
}
