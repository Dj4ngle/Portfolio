package posts

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"redditclone/internal/user"
	"time"
)

const (
	ErrBadRequest    = `{"message": "bad request"}`
	ErrPostNotFound  = `{"message": "post not found"}`
	ErrFailedUpdate  = `{"message": "failed to update field"}`
	ErrFailedConvert = `{"message": "failed to convert values"}`
	ErrFailedDelete  = `{"message": "failed to delete"}`
)

type PostMongoRepository struct {
	DB *mongo.Collection
}

func NewMongoRepo(db *mongo.Collection) *PostMongoRepository {
	return &PostMongoRepository{DB: db}
}

// Функции фильтрации (попытался в оптимизацию кода)
func FilterByCategory(category string) func(*Post) bool {
	return func(p *Post) bool {
		return p.Category == category
	}
}

func FilterByUser(user string) func(*Post) bool {
	return func(p *Post) bool {
		return p.Author.Username == user
	}
}

func FilterAll() func(*Post) bool {
	return func(p *Post) bool {
		return true
	}
}

func (repo *PostMongoRepository) GetPost(postID primitive.ObjectID) (*Post, error) {
	var post *Post

	err := repo.DB.FindOneAndUpdate(
		context.Background(),
		bson.D{
			{Key: "_id", Value: postID},
		},
		bson.D{{Key: "$inc", Value: bson.M{"views": 1}}},
		options.FindOneAndUpdate().SetReturnDocument(1),
	).Decode(&post)

	if err != nil {
		return nil, errors.New(ErrPostNotFound)
	}

	return post, nil
}

func (repo *PostMongoRepository) GetPosts(filter func(*Post) bool) ([]*Post, error) {
	var posts, newPosts []*Post

	c, err := repo.DB.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, errors.New(ErrPostNotFound)
	}

	err = c.All(context.Background(), &posts)
	if err != nil {
		return nil, errors.New(ErrFailedConvert)
	}

	for _, v := range posts {
		if filter(v) {
			newPosts = append(newPosts, v)
		}
	}

	return newPosts, nil
}

func (repo *PostMongoRepository) VotePost(postID primitive.ObjectID, user int64, voteVal int) (*Post, error) {
	post, err := repo.GetPost(postID)
	if err != nil {
		return nil, errors.New(ErrPostNotFound)
	}

	index := -1
	for k, v := range post.Votes {
		if v.UserID == user {
			if v.Vote == voteVal {
				return post, nil
			}
			index = k
			post.Score += voteVal - v.Vote
			post.Votes[k].Vote = voteVal
			post.UpvoteCount += voteVal
			continue
		}
	}

	if index == -1 {
		vote := Vote{
			UserID: user,
			Vote:   voteVal,
		}
		post.UpvoteCount += voteVal
		post.Score += voteVal
		post.VoteCount++
		post.Votes = append(post.Votes, &vote)
	}

	if post.VoteCount > 0 {
		post.UpvotePercentage = post.UpvoteCount * 100 / post.VoteCount
	} else {
		post.UpvotePercentage = 0
	}

	// Создание фильтра по ID
	filter := bson.M{"_id": post.ID}
	// Замена существующего документа
	result, err := repo.DB.ReplaceOne(context.Background(), filter, post)
	if err != nil {
		return nil, errors.New(ErrBadRequest)
	}
	if result.ModifiedCount == 0 {
		return nil, errors.New(ErrFailedUpdate)
	}

	return post, nil
}

func (repo *PostMongoRepository) UnVotePost(postID primitive.ObjectID, user int64) (*Post, error) {
	post, err := repo.GetPost(postID)
	if err != nil {
		return nil, errors.New(ErrPostNotFound)
	}

	index := -1
	for k, v := range post.Votes {
		if v.UserID == user {
			if v.Vote == 1 {
				post.UpvoteCount--
				post.Score--
			} else {
				post.Score++
			}
			index = k
			post.VoteCount--
			post.Votes = append(post.Votes[:k], post.Votes[k+1:]...)
		}
	}
	if index == -1 {
		return nil, errors.New(ErrFailedUpdate)
	}

	if post.VoteCount > 0 {
		post.UpvotePercentage = post.UpvoteCount * 100 / post.VoteCount
	} else {
		post.UpvotePercentage = 0
	}

	// Создание фильтра по ID
	filter := bson.M{"_id": post.ID}
	// Замена существующего документа
	result, err := repo.DB.ReplaceOne(context.Background(), filter, post)
	if err != nil {
		return nil, errors.New(ErrFailedUpdate)
	}
	if result.ModifiedCount == 0 {
		return nil, errors.New(ErrFailedUpdate)
	}

	return post, nil
}

func (repo *PostMongoRepository) MakePost(newPost *PostForm, username string, userID int64) (*Post, error) {
	var post *Post

	switch newPost.Type {
	case "text":
		post = &Post{
			ID:       primitive.NewObjectID(),
			Title:    newPost.Title,
			Category: newPost.Category,
			Type:     newPost.Type,
			Text:     newPost.Text,
			Created:  time.Now().UTC().Format(time.RFC3339),
			Author: &user.User{
				ID:       userID,
				Username: username,
			},
			Votes: []*Vote{
				{
					UserID: userID,
					Vote:   1,
				},
			},
			Score:            1,
			UpvotePercentage: 100,
			Comments:         []*Comment{},
		}
	case "link":
		post = &Post{
			ID:       primitive.NewObjectID(),
			Title:    newPost.Title,
			Category: newPost.Category,
			Type:     newPost.Type,
			Text:     newPost.Text,
			Created:  time.Now().UTC().Format(time.RFC3339),
			Author: &user.User{
				ID:       userID,
				Username: username,
			},
			Votes: []*Vote{
				{
					UserID: userID,
					Vote:   1,
				},
			},
			Score:            1,
			UpvotePercentage: 100,
			Comments:         []*Comment{},
		}
	default:
		return nil, errors.New(ErrBadRequest)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := repo.DB.InsertOne(ctx, post)
	if err != nil {
		return nil, errors.New(ErrBadRequest)
	}

	return post, nil
}

func (repo *PostMongoRepository) DeletePost(postID primitive.ObjectID, userID int64) (bool, error) {
	filter := bson.M{"_id": postID, "author.id": userID}
	result, err := repo.DB.DeleteOne(context.Background(), filter)
	if err != nil {
		return false, errors.New(ErrBadRequest)
	}
	if result.DeletedCount == 0 {
		return false, errors.New(ErrFailedDelete)
	}

	return true, nil
}

func (repo *PostMongoRepository) MakeComment(postID primitive.ObjectID, comment, username string, userID int64) (*Post, error) {
	newComment := &Comment{
		ID: primitive.NewObjectID(),
		Author: &user.User{
			ID:       userID,
			Username: username,
		},
		Body:    comment,
		Created: time.Now().UTC().Format(time.RFC3339),
	}

	filter := bson.M{"_id": postID}
	update := bson.M{"$push": bson.M{"comments": newComment}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	result := repo.DB.FindOneAndUpdate(context.Background(), filter, update, opts)
	if result.Err() != nil {
		return nil, errors.New(ErrBadRequest)
	}

	var post Post
	err := result.Decode(&post)
	if err != nil {
		return nil, errors.New(ErrBadRequest)
	}

	return &post, nil
}

func (repo *PostMongoRepository) DeleteComment(postID primitive.ObjectID, commentID primitive.ObjectID, userID int64) (*Post, error) {
	filter := bson.M{"_id": postID, "comments": bson.M{"$elemMatch": bson.M{"_id": commentID, "author.id": userID}}}
	update := bson.M{"$pull": bson.M{"comments": bson.M{"_id": commentID}}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	result := repo.DB.FindOneAndUpdate(context.Background(), filter, update, opts)
	if result.Err() != nil {
		return nil, errors.New(ErrFailedUpdate)
	}

	var post Post
	err := result.Decode(&post)
	if err != nil {
		return nil, errors.New(ErrBadRequest)
	}

	return &post, nil
}
