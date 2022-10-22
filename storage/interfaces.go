package storage

import "context"

type Post struct {
	Id        string `json:"id"`
	Text      string `json:"text"`
	AuthorId  string `json:"authorId"`
	CreatedAt string `json:"createdAt"`
}

type Storage interface {
	AddPost(ctx context.Context, post Post) error
	GetPost(ctx context.Context, postId string) (Post, error)
	GetPostsByUser(ctx context.Context, userId string, page string, size int) ([]Post, error)
}
