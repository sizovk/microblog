package inmemory

import (
	"context"
	"microblog/storage"
	"sync"
)

func NewStorage() *InmemoryStorage {
	return &InmemoryStorage{
		postIdToPost:    make(map[string]storage.Post),
		userIdToPostIds: make(map[string][]string),
		pageIdToOffset:  make(map[string]int),
	}
}

type InmemoryStorage struct {
	mu              sync.RWMutex
	postIdToPost    map[string]storage.Post
	userIdToPostIds map[string][]string
	pageIdToOffset  map[string]int
}

func (is *InmemoryStorage) AddPost(ctx context.Context, post storage.Post) error {
	is.mu.Lock()
	defer is.mu.Unlock()
	if _, found := is.postIdToPost[post.Id]; found {
		return storage.ErrCollision
	}

	is.postIdToPost[post.Id] = post
	is.userIdToPostIds[post.AuthorId] = append(is.userIdToPostIds[post.AuthorId], post.Id)
	return nil
}

func (is *InmemoryStorage) GetPost(ctx context.Context, postId string) (storage.Post, error) {
	is.mu.Lock()
	defer is.mu.Unlock()
	if _, found := is.postIdToPost[postId]; !found {
		return storage.Post{}, storage.ErrPostNotFound
	}
	return is.postIdToPost[postId], nil
}

func (is *InmemoryStorage) GetPostsByUser(ctx context.Context, userId string, page string, size int) (storage.UserPosts, error) {
	is.mu.Lock()
	defer is.mu.Unlock()
	postIds, found := is.userIdToPostIds[userId]
	if !found {
		if page == "" {
			return storage.UserPosts{Posts: []storage.Post{}}, nil
		}
		return storage.UserPosts{Posts: []storage.Post{}}, storage.ErrUserNotFound
	}
	startInd := len(postIds) - 1
	if page != "" {
		startInd, found = is.pageIdToOffset[page]
		if !found {
			return storage.UserPosts{Posts: []storage.Post{}}, storage.ErrWrongPage
		}
		if is.postIdToPost[page].AuthorId != userId {
			return storage.UserPosts{Posts: []storage.Post{}}, storage.ErrWrongAuthor
		}
	}
	lastInd := startInd - size
	nextPage := ""
	if lastInd < 0 {
		lastInd = -1
	} else {
		nextPage = postIds[lastInd]
		is.pageIdToOffset[nextPage] = lastInd
	}
	posts := make([]storage.Post, startInd-lastInd)
	for i := range posts {
		posts[i] = is.postIdToPost[postIds[startInd-i]]
	}
	return storage.UserPosts{Posts: posts, NextPage: nextPage}, nil
}

func (is *InmemoryStorage) PatchPost(ctx context.Context, post storage.Post) error {
	is.mu.Lock()
	defer is.mu.Unlock()

	is.postIdToPost[post.Id] = post
	return nil
}
