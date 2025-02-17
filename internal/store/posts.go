package store

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
)

type Post struct {
	ID int				`json:"id"`
	Title string		`json:"title"`
	Content string		`json:"content"`
	UserId int64		`json:"user_id"`
	Tags []string		`json:"tags"`
	CreatedAt string	`json:"created_at"`
	UpdatedAt string	`json:"updated_at"`
	Comments []Comment	`json:"comments"`
}

type PostsStore struct {
	db *sql.DB
}

func (s *PostsStore) Create(ctx context.Context, post *Post) error{
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
	query := `SELECT id, title, content, user_id, tags, created_at, updated_at FROM posts WHERE id = $1`

	post := &Post{}

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.UserId,
		pq.Array(&post.Tags),
		&post.CreatedAt,
		&post.UpdatedAt,
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
	query := `UPDATE posts SET title = $1, content = $2, tags = $3, updated_at = now() WHERE id = $4`

	_, err := s.db.ExecContext(ctx, query, post.Title, post.Content, pq.Array(post.Tags), post.ID)

	if err != nil {
		return err
	}

	return nil
}