package posts

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"redditclone/internal/user"
)

type Comment struct {
	Created string             `json:"created"`
	Author  *user.User         `json:"author"`
	Body    string             `json:"body"`
	ID      primitive.ObjectID `json:"id" bson:"_id"`
}

type Vote struct {
	UserID int64 `json:"user"`
	Vote   int   `json:"vote"`
}

type Post struct {
	ID               primitive.ObjectID `json:"id" bson:"_id"`
	Score            int                `json:"score"`
	Views            int                `json:"views"`
	Type             string             `json:"type"`
	Title            string             `json:"title"`
	Author           *user.User         `json:"author"`
	Category         string             `json:"category"`
	Text             string             `json:"text,omitempty" bson:"text,omitempty"`
	Votes            []*Vote            `json:"votes"`
	Comments         []*Comment         `json:"comments"`
	Created          string             `json:"created"`
	UpvotePercentage int                `json:"upvotePercentage"`
	UpvoteCount      int                `json:"upvotecount"`
	VoteCount        int                `json:"votecount"`
	URL              string             `json:"url,omitempty" bson:"url,omitempty"`
}

type PostForm struct {
	Type     string `json:"type"  validate:"required"`
	Title    string `json:"title"  validate:"required"`
	Category string `json:"category"  validate:"required"`
	Text     string `json:"text,omitempty"`
	URL      string `json:"url,omitempty"`
}

type CommentForm struct {
	Body string `json:"comment"  validate:"required"`
}

//go:generate mockgen -source=posts.go -destination=repo_mock.go -package=posts PostRepo
type PostRepo interface {
	GetPost(postID primitive.ObjectID) (*Post, error)
	GetPosts(filter func(*Post) bool) ([]*Post, error)
	VotePost(postID primitive.ObjectID, user int64, voteVal int) (*Post, error)
	UnVotePost(postID primitive.ObjectID, user int64) (*Post, error)
	MakePost(newPost *PostForm, username string, userID int64) (*Post, error)
	DeletePost(postID primitive.ObjectID, userID int64) (bool, error)
	MakeComment(postID primitive.ObjectID, comment, username string, userID int64) (*Post, error)
	DeleteComment(postID primitive.ObjectID, commentID primitive.ObjectID, userID int64) (*Post, error)
}
