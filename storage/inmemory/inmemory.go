package inmemory

import (
	"context"
	"microblog/storage"
)

func NewStorage() *InmemoryStorage {
	return &InmemoryStorage{}
}

type InmemoryStorage struct {
}

func (is *InmemoryStorage) AddPost(ctx context.Context, post storage.Post) error {
	return nil
}

func (is *InmemoryStorage) GetPost(ctx context.Context, postId string) (storage.Post, error) {
	return storage.Post{}, nil
}

func (is *InmemoryStorage) GetPostsByUser(ctx context.Context, userId string, page string, size int) ([]storage.Post, error) {
	return []storage.Post{}, nil
}
