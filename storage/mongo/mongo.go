package mongo

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"microblog/storage"
	"time"
)

const collName = "posts"

func NewStorage(mongoURL string, dbName string) *MongoStorage {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
	if err != nil {
		panic(err)
	}

	collection := client.Database(dbName).Collection(collName)
	ensureIndexes(ctx, collection)

	return &MongoStorage{
		posts: collection,
	}
}

func ensureIndexes(ctx context.Context, collection *mongo.Collection) {
	indexModels := []mongo.IndexModel{
		{
			Keys: bsonx.Doc{
				{Key: "authorId", Value: bsonx.Int32(1)},
				{Key: "_id", Value: bsonx.Int32(-1)},
			},
		},
	}
	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)

	_, err := collection.Indexes().CreateMany(ctx, indexModels, opts)
	if err != nil {
		panic(fmt.Errorf("failed to ensure indexes %w", err))
	}
}

type MongoStorage struct {
	posts *mongo.Collection
}

func (ms *MongoStorage) AddPost(ctx context.Context, post storage.Post) error {
	_, err := ms.posts.InsertOne(ctx, post)
	if mongo.IsDuplicateKeyError(err) {
		return storage.ErrCollision
	}
	if err != nil {
		return err
	}
	return nil
}

func (ms *MongoStorage) GetPost(ctx context.Context, postId string) (storage.Post, error) {
	var post storage.Post
	err := ms.posts.FindOne(ctx, bson.M{"_id": postId}).Decode(&post)
	if err != nil {
		return storage.Post{}, storage.ErrPostNotFound
	}
	return post, nil
}

func (ms *MongoStorage) GetPostsByUser(ctx context.Context, userId string, page string, size int) (storage.UserPosts, error) {
	posts := []storage.Post{}
	opt := options.Find().SetSort(bson.D{{"authorId", 1}, {"_id", -1}}).SetLimit(int64(size) + 1)
	var cursor *mongo.Cursor
	var err error
	if page == "" {
		cursor, err = ms.posts.Find(ctx, bson.M{"authorId": userId}, opt)
		if err != nil {
			return storage.UserPosts{Posts: posts}, nil
		}
	} else {
		num, err := ms.posts.CountDocuments(ctx, bson.M{"authorId": userId, "_id": page})
		if num == 0 || err != nil {
			return storage.UserPosts{Posts: posts}, storage.ErrWrongPage
		}
		cursor, err = ms.posts.Find(ctx, bson.M{"authorId": userId, "_id": bson.M{"$lt": page}}, opt)
		if err != nil {
			return storage.UserPosts{Posts: posts}, storage.ErrUserNotFound
		}
	}
	var post storage.Post
	for cursor.Next(ctx) {
		err := cursor.Decode(&post)
		if err != nil {
			return storage.UserPosts{Posts: posts}, err
		}
		posts = append(posts, post)
	}
	if len(posts) == size+1 {
		posts = posts[:len(posts)-1]
		return storage.UserPosts{Posts: posts, NextPage: posts[len(posts)-1].Id}, nil
	} else {
		return storage.UserPosts{Posts: posts}, nil
	}
}

func (ms *MongoStorage) PatchPost(ctx context.Context, post storage.Post) error {
	update := bson.D{
		{"$set", bson.M{"text": post.Text}},
		{"$set", bson.M{"lastModifiedAt": post.LastModifiedAt}},
	}
	_, err := ms.posts.UpdateByID(ctx, post.Id, update)
	if err != nil {
		return err
	}
	return nil
}
