package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

type Post struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	UserId    int64     `json:"user_id"`
	Tags      []string  `json:"tags"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	Comments  []Comment `json:"comments"`
	Version   int       `json:"version"`
	User      User      `json:"user"`
}

type PostWithMetadata struct {
	Post

	CommentsCount int `json:"comments_count"`
}

type PostsStore struct {
	db *sql.DB
}

func (s *PostsStore) Create(ctx context.Context, post *Post) error {
	// Create a new post
	query := `INSERT INTO posts (title, content, user_id, tags) VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at`

	err := s.db.QueryRowContext(ctx,
		query,
		post.Title,
		post.Content,
		post.UserId,
		pq.Array(post.Tags),
	).Scan(
		&post.ID,
		&post.CreatedAt,
		&post.UpdatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (s *PostsStore) GetById(ctx context.Context, id int) (*Post, error) {
	// Get post by id
	query := `SELECT id, title, content, user_id, tags, created_at, updated_at, version FROM posts WHERE id = $1`

	post := &Post{}

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.UserId,
		pq.Array(&post.Tags),
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.Version,
	)

	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return post, nil
}

func (s *PostsStore) Delete(ctx context.Context, id int) error {
	// Delete post by id
	query := `DELETE FROM posts WHERE id = $1`

	res, err := s.db.ExecContext(ctx, query, id)

	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()

	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *PostsStore) Update(ctx context.Context, post *Post) error {
	// Update post
	query := `
		UPDATE posts 
		SET title = $1, content = $2, updated_at = now(), version = version + 1 
		WHERE id = $3 AND version = $4
		RETURNING version`

	err := s.db.QueryRowContext(
		ctx,
		query,
		post.Title,
		post.Content,
		post.ID,
		post.Version,
	).Scan(&post.Version)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrNotFound
		default:
			return err
		}
	}

	return nil
}

func (s *PostsStore) GetUserFeed(ctx context.Context, userId int64, fq PaginatedFieldQuery) ([]PostWithMetadata, error) {
	//TODO : implement time sorting

	baseQuery := `
        SELECT p.id, p.user_id, p.title, p.content, p.created_at, p.version, p.tags,
           u.username, u.email,
           COUNT(c.id) as comments_count
        FROM posts p
        LEFT JOIN comments c ON p.id = c.post_id
        LEFT JOIN users u ON p.user_id = u.id
        JOIN followers f ON p.user_id = f.follower_id OR p.user_id = $1
        WHERE f.user_id = $1`

	if fq.Search != "" {
		baseQuery += ` AND (p.title ILIKE '%' || $4 || '%' OR p.content ILIKE '%' || $4 || '%')`
	}

	if len(fq.Tags) > 0 {
		baseQuery += ` AND (p.tags @> $5)`
	}

	baseQuery += `
        GROUP BY p.id, p.user_id, p.title, p.content, p.created_at, p.version, p.tags, 
        u.username, u.email
        ORDER BY p.created_at ` + fq.Sort + `
        LIMIT $2 OFFSET $3`

	var rows *sql.Rows
	var err error

	if fq.Search != "" && len(fq.Tags) > 0 {
		rows, err = s.db.QueryContext(ctx, baseQuery, userId, fq.Limit, fq.Offset, fq.Search, pq.Array(fq.Tags))
	} else if fq.Search != "" {
		rows, err = s.db.QueryContext(ctx, baseQuery, userId, fq.Limit, fq.Offset, fq.Search)
	} else if len(fq.Tags) > 0 {
		rows, err = s.db.QueryContext(ctx, baseQuery, userId, fq.Limit, fq.Offset, pq.Array(fq.Tags))
	} else {
		rows, err = s.db.QueryContext(ctx, baseQuery, userId, fq.Limit, fq.Offset)
	}

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	defer rows.Close()

	posts := []PostWithMetadata{}

	for rows.Next() {
		post := PostWithMetadata{}

		err := rows.Scan(
			&post.ID,
			&post.UserId,
			&post.Title,
			&post.Content,
			&post.CreatedAt,
			&post.Version,
			pq.Array(&post.Tags),
			&post.User.Username,
			&post.User.Email,
			&post.CommentsCount,
		)

		if err != nil {
			switch err {
			case sql.ErrNoRows:
				return nil, ErrNotFound
			default:
				return nil, err
			}
		}

		posts = append(posts, post)
	}

	return posts, nil
}
