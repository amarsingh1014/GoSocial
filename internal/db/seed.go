package db

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"social/internal/store"
	"strconv"
	"time"
)

var names = []string{
	"John", "Jane", "Alice", "Bob", "Charlie", "Dave", "Eve", "Frank", "Grace", "Hank",
	"Ivy", "Jack", "Kathy", "Leo", "Mona", "Nate", "Olivia", "Paul", "Quincy", "Rachel",
	"Sam", "Tina", "Uma", "Victor", "Wendy", "Xander", "Yara", "Zane",
}

var titles = []string{
	"First Post", "Hello World", "My Journey", "Tech Talk", "Life Lessons", "Travel Diaries", "Food Adventures", "Book Reviews", "Movie Critiques", "Daily Thoughts",
	"Random Thoughts", "Tech Innovations", "Life Hacks", "Travel Tips", "Food Recipes", "Book Summaries", "Movie Reviews", "Daily Routines", "Fitness Tips", "Coding Tutorials",
}

var tags = []string{
	"tech", "life", "travel", "food", "books", "movies", "thoughts", "adventure", "journey", "lessons",
	"fitness", "coding", "recipes", "tips", "reviews", "summaries", "hacks", "routines", "innovations", "diaries",
}

func Seed(store *store.Storage) error {
	ctx := context.Background()

	// Seed users
	users := generateUsers(100)
	for i, user := range users {
		// Generate a token for each user
		token := fmt.Sprintf("token-for-user-%d", i)
		expiry := time.Hour * 24 // 1 day expiration

		if err := store.Users.CreateAndInvite(ctx, &user, token, expiry); err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		users[i].ID = user.ID // Ensure the ID is set correctly
		log.Printf("Created user: %s with ID: %d", user.Username, user.ID)
	}

	// Seed posts
	posts := generatePosts(users, 200)
	for _, post := range posts {
		if err := store.Posts.Create(ctx, &post); err != nil {
			return fmt.Errorf("failed to create post: %w", err)
		}
		log.Printf("Created post: %s", post.Title)

		// Seed comments for each post
		comments := generateComments(users, post, 10)
		for _, comment := range comments {
			if err := store.Comments.Create(ctx, &comment); err != nil {
				return fmt.Errorf("failed to create comment: %w", err)
			}
			log.Printf("Created comment: %s", comment.Content)
		}
	}

	return nil
}

func generateUsers(num int) []store.User {
	users := make([]store.User, 0, num)

	for i := 0; i < num; i++ {
		name := names[i%len(names)]
		users = append(users, store.User{
			Username: name + strconv.Itoa(i),
			Email:    name + strconv.Itoa(i) + "@example.com",
			RoleID: 1,
		})
	}

	return users
}

func generatePosts(users []store.User, num int) []store.Post {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	posts := make([]store.Post, 0, num)

	for i := 0; i < num; i++ {
		user := users[i%len(users)]
		title := titles[r.Intn(len(titles))]
		content := fmt.Sprintf("This is the content of post %d by %s", i, user.Username)
		postTags := generateRandomTags(r, 3)
		posts = append(posts, store.Post{
			Title:   title,
			Content: content,
			UserId:  user.ID,
			Tags:    postTags,
		})
	}

	return posts
}

func generateRandomTags(r *rand.Rand, num int) []string {
	selectedTags := make([]string, 0, num)
	for i := 0; i < num; i++ {
		tag := tags[r.Intn(len(tags))]
		selectedTags = append(selectedTags, tag)
	}
	return selectedTags
}

func generateComments(users []store.User, post store.Post, num int) []store.Comment {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	comments := make([]store.Comment, 0, num)

	for i := 0; i < num; i++ {
		user := users[r.Intn(len(users))]
		content := fmt.Sprintf("This is a comment %d by %s on post %d", i, user.Username, post.ID)
		comments = append(comments, store.Comment{
			PostID:  post.ID,
			UserID:  (int)(user.ID),
			Content: content,
		})
	}

	return comments
}
