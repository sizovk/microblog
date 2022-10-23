package storage

import (
	"context"
	"errors"
)

type Post struct {
	Id        string `json:"id"`
	Text      string `json:"text"`
	AuthorId  string `json:"authorId"`
	CreatedAt string `json:"createdAt"`
}

type UserPosts struct {
	Posts    []Post `json:"posts"`
	NextPage string `json:"nextPage,omitempty"`
}

var (
	ErrCollision    = errors.New("collision_error")
	ErrPostNotFound = errors.New("post_not_found")
	ErrUserNotFound = errors.New("user_not_found")
)

type Storage interface {
	AddPost(ctx context.Context, post Post) error
	GetPost(ctx context.Context, postId string) (Post, error)
	GetPostsByUser(ctx context.Context, userId string, page string, size int) (UserPosts, error)
}
