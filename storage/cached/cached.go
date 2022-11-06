package cached

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"microblog/storage"
	"time"
)

const redisExpiration = time.Hour
const redisKeyPrefix = "microblog:"

func NewStorage(client *redis.Client, persistentStorage storage.Storage) *CachedStorage {
	return &CachedStorage{client: client, persistentStorage: persistentStorage}
}

type CachedStorage struct {
	client            *redis.Client
	persistentStorage storage.Storage
}

func (cs *CachedStorage) AddPost(ctx context.Context, post storage.Post) error {
	err := cs.persistentStorage.AddPost(ctx, post)
	if err != nil {
		return err
	}

	if err := cs.storePost(ctx, post); err != nil {
		return err
	}
	return nil
}

func (cs *CachedStorage) GetPost(ctx context.Context, postId string) (storage.Post, error) {
	var post storage.Post
	post, err := cs.restorePost(ctx, postId)
	if err == nil {
		return post, nil
	}
	if err != redis.Nil {
		return storage.Post{}, err
	}
	post, err = cs.persistentStorage.GetPost(ctx, postId)
	if err != nil {
		return storage.Post{}, err
	}
	if err := cs.storePost(ctx, post); err != nil {
		return post, err
	}
	return post, nil
}

func (cs *CachedStorage) GetPostsByUser(ctx context.Context, userId string, page string, size int) (storage.UserPosts, error) {
	return cs.persistentStorage.GetPostsByUser(ctx, userId, page, size)
}

func (cs *CachedStorage) PatchPost(ctx context.Context, post storage.Post) error {
	if err := cs.persistentStorage.PatchPost(ctx, post); err != nil {
		return err
	}
	if err := cs.storePost(ctx, post); err != nil {
		return err
	}
	return nil
}

func (cs *CachedStorage) restorePost(ctx context.Context, postId string) (storage.Post, error) {
	rawPost, err := cs.client.Get(ctx, cs.redisKey(postId)).Result()
	if err != nil {
		return storage.Post{}, err
	}
	var post storage.Post
	err = json.Unmarshal([]byte(rawPost), &post)
	return post, err
}

func (cs *CachedStorage) storePost(ctx context.Context, post storage.Post) error {
	rawPost, err := json.Marshal(post)
	if err != nil {
		return err
	}
	err = cs.client.Set(ctx, cs.redisKey(post.Id), rawPost, redisExpiration).Err()
	return err
}

func (m *CachedStorage) redisKey(shortKey string) string {
	// add a prefix not to collide with other data stored in the same redis
	return redisKeyPrefix + shortKey
}
