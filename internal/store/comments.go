package store

import (
	"context"
	"database/sql"
)

type Comment struct {
	ID        int    `json:"id"`
	PostID    int    `json:"post_id"`
	UserID    int    `json:"user_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	User      User `json:"user"`	
}

type CommentsStore struct {
	db *sql.DB
}

func (s * CommentsStore) GetByPostID(ctx context.Context, postID int) ([]Comment, error) {
	// Get comments by post id
	query := `SELECT c.id, c.post_id, c.user_id, c.content, c.created_at, users.id, users.username FROM comments c
		JOIN users ON users.id = c.user_id
		WHERE c.post_id = $1
		ORDER BY c.created_at DESC`

	rows, err := s.db.QueryContext(ctx, query, postID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	comments := []Comment{}

	for rows.Next() {
		comment := Comment{}

		comment.User = User{}

		err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.UserID,
			&comment.Content,
			&comment.CreatedAt,
			&comment.User.ID,
			&comment.User.Username,
		)

		if err != nil {
			return nil, err
		}

		comments = append(comments, comment)
	}

	return comments, nil
}