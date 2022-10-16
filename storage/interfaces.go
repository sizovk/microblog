package storage

import "context"

type Post struct {
	Id        string `json:"id"`
	Text      string `json:"text"`
	AuthorId  string `json:"authorId"`
	CreatedAt string `json:"createdAt"`
}

type Storage interface {
	CreatePost(ctx context.Context) (Post, error)
	GetPost(ctx context.Context, postId string) (Post, error)
	GetPostsByUser(ctx context.Context, userId string) ([]Post, error)
}
