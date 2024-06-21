package posts

import (
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
	"redditclone/internal/user"
	"testing"
)

func TestGetPost(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	postID := primitive.NewObjectID()

	var tests = []struct {
		name          string
		mockResponses []bson.D
		postID        primitive.ObjectID
		expectedPost  *Post
		expectError   bool
	}{
		{
			name: "Проверка на успешное выполнение запроса для получения поста",
			mockResponses: []bson.D{{
				{Key: "ok", Value: 1},
				{Key: "value", Value: bson.D{
					{Key: "_id", Value: postID},
					{Key: "Score", Value: 0},
					{Key: "Views", Value: 9},
					{Key: "Type", Value: "text"},
					{Key: "Title", Value: "Post 1"},
					{Key: "Category", Value: "fashion"},
					{Key: "Text", Value: "Text of the post 1"},
					{Key: "Created", Value: "2024-05-07T20:38:17Z"},
					{Key: "UpvotePercentage", Value: 0},
				}},
			}},
			postID: primitive.NewObjectID(),
			expectedPost: &Post{
				ID:               postID,
				Score:            0,
				Views:            9,
				Type:             "text",
				Title:            "Post 1",
				Category:         "fashion",
				Text:             "Text of the post 1",
				Created:          "2024-05-07T20:38:17Z",
				UpvotePercentage: 0,
			},
			expectError: false,
		},
		{
			name:          "Проверка на обработку ошибки при запросе",
			mockResponses: []bson.D{{{Key: "ok", Value: 0}}},
			postID:        postID,
			expectedPost:  nil,
			expectError:   true,
		},
	}

	for _, tc := range tests {
		mt.Run(tc.name, func(mt *mtest.T) {
			repo := NewMongoRepo(mt.Coll)
			mt.ClearMockResponses()
			mt.AddMockResponses(tc.mockResponses...)

			post, err := repo.GetPost(tc.postID)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, post)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedPost, post)
			}
		})
	}
}

func TestGetPosts(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	postID1 := primitive.NewObjectID()
	postID2 := primitive.NewObjectID()
	postID3 := primitive.NewObjectID()
	expectedPosts := []*Post{
		{
			ID:               postID1,
			Score:            0,
			Views:            9,
			Type:             "text",
			Title:            "Post 1",
			Category:         "fashion",
			Text:             "Text of the post 1",
			Created:          "2024-05-07T20:38:17Z",
			UpvotePercentage: 0,
			Author: &user.User{
				ID:       1,
				Username: "vasya",
			},
		},
		{
			ID:               postID2,
			Score:            0,
			Views:            9,
			Type:             "text",
			Title:            "Post 2",
			Category:         "music",
			Text:             "Text of the post 1",
			Created:          "2024-05-07T20:38:17Z",
			UpvotePercentage: 0,
			Author: &user.User{
				ID:       2,
				Username: "dima",
			},
		},
		{
			ID:               postID3,
			Score:            0,
			Views:            9,
			Type:             "text",
			Title:            "Post 3",
			Category:         "fashion",
			Text:             "Text of the post 1",
			Created:          "2024-05-07T20:38:17Z",
			UpvotePercentage: 0,
			Author: &user.User{
				ID:       3,
				Username: "ivan",
			},
		},
	}

	first := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
		{Key: "_id", Value: postID1},
		{Key: "Score", Value: 0},
		{Key: "Views", Value: 9},
		{Key: "Type", Value: "text"},
		{Key: "Title", Value: "Post 1"},
		{Key: "Category", Value: "fashion"},
		{Key: "Text", Value: "Text of the post 1"},
		{Key: "Created", Value: "2024-05-07T20:38:17Z"},
		{Key: "UpvotePercentage", Value: 0},
		{Key: "Author", Value: bson.M{
			"id":       1,
			"username": "vasya",
		}},
	})
	second := mtest.CreateCursorResponse(1, "foo.bar", mtest.NextBatch, bson.D{
		{Key: "_id", Value: postID2},
		{Key: "Score", Value: 0},
		{Key: "Views", Value: 9},
		{Key: "Type", Value: "text"},
		{Key: "Title", Value: "Post 2"},
		{Key: "Category", Value: "music"},
		{Key: "Text", Value: "Text of the post 1"},
		{Key: "Created", Value: "2024-05-07T20:38:17Z"},
		{Key: "UpvotePercentage", Value: 0},
		{Key: "Author", Value: bson.M{
			"id":       2,
			"username": "dima",
		}},
	})
	third := mtest.CreateCursorResponse(1, "foo.bar", mtest.NextBatch, bson.D{
		{Key: "_id", Value: postID3},
		{Key: "Score", Value: 0},
		{Key: "Views", Value: 9},
		{Key: "Type", Value: "text"},
		{Key: "Title", Value: "Post 3"},
		{Key: "Category", Value: "fashion"},
		{Key: "Text", Value: "Text of the post 1"},
		{Key: "Created", Value: "2024-05-07T20:38:17Z"},
		{Key: "UpvotePercentage", Value: 0},
		{Key: "Author", Value: bson.M{
			"id":       3,
			"username": "ivan",
		}},
	})

	killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)

	testCases := []struct {
		name          string
		filter        func(*Post) bool
		mockResponses []bson.D
		expectedPosts []*Post
		expectedError string
	}{
		{
			name:          "Проверка на успешное выполнение запроса для получения всех постов",
			filter:        FilterAll(),
			mockResponses: []bson.D{first, second, third, killCursors},
			expectedPosts: expectedPosts,
		},
		{
			name:          "Проверка на успешное выполнение запроса для постов по категории",
			filter:        FilterByCategory("music"),
			mockResponses: []bson.D{first, second, third, killCursors},
			expectedPosts: []*Post{expectedPosts[1]},
		},
		{
			name:          "Проверка на успешное выполнение запроса для постов по пользователю",
			filter:        FilterByUser("ivan"),
			mockResponses: []bson.D{first, second, third, killCursors},
			expectedPosts: []*Post{expectedPosts[2]},
		},
		{
			name:          "Проверка на обработку ошибки при запросе",
			filter:        FilterAll(),
			mockResponses: []bson.D{{{Key: "ok", Value: 0}}},
			expectedPosts: nil,
			expectedError: ErrPostNotFound,
		},
		{
			name:          "Проверка на обработку ошибки при конвертации всех постов",
			filter:        FilterAll(),
			mockResponses: []bson.D{mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{{Key: "_id", Value: "postID3"}, {Key: "Score", Value: "0"}}), killCursors},
			expectedPosts: nil,
			expectedError: ErrFailedConvert,
		},
	}

	for _, tc := range testCases {
		mt.Run(tc.name, func(mt *mtest.T) {
			repo := NewMongoRepo(mt.Coll)
			mt.ClearMockResponses()
			mt.AddMockResponses(tc.mockResponses...)

			posts, err := repo.GetPosts(tc.filter)

			if err != nil {
				assert.Equal(t, tc.expectedError, err.Error())
				assert.Nil(t, posts)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedPosts, posts)
			}
		})
	}
}

func TestVotePost(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	postID := primitive.NewObjectID()
	post := bson.D{
		{Key: "ok", Value: 1},
		{Key: "value", Value: bson.D{
			{Key: "_id", Value: postID},
			{Key: "Score", Value: 1},
			{Key: "Views", Value: 9},
			{Key: "Type", Value: "text"},
			{Key: "Title", Value: "Post 3"},
			{Key: "Category", Value: "fashion"},
			{Key: "Text", Value: "Text of the post 1"},
			{Key: "Created", Value: "2024-05-07T20:38:17Z"},
			{Key: "UpvotePercentage", Value: 100},
			{Key: "Upvotecount", Value: 2},
			{Key: "Votecount", Value: 3},
			{Key: "Author", Value: bson.M{
				"id":       3,
				"username": "ivan",
			}},
			{Key: "Votes", Value: bson.A{
				bson.D{{Key: "userID", Value: 1}, {Key: "vote", Value: 1}},
				bson.D{{Key: "userID", Value: 2}, {Key: "vote", Value: 1}},
				bson.D{{Key: "userID", Value: 3}, {Key: "vote", Value: -1}},
			}},
		}},
	}
	badPost := bson.D{
		{Key: "ok", Value: 1},
		{Key: "value", Value: bson.D{
			{Key: "_id", Value: postID},
			{Key: "Score", Value: 1},
			{Key: "Views", Value: 3},
			{Key: "Type", Value: "text"},
			{Key: "Title", Value: "Post 1"},
			{Key: "Category", Value: "fashion"},
			{Key: "Text", Value: "Text of the post 1"},
			{Key: "Created", Value: "2024-05-07T20:38:17Z"},
			{Key: "UpvotePercentage", Value: 0},
			{Key: "Upvotecount", Value: 0},
			{Key: "Votecount", Value: 0},
			{Key: "Author", Value: bson.M{
				"id":       1,
				"username": "ivan",
			}},
			{Key: "Votes", Value: bson.A{
				bson.D{{Key: "userID", Value: 1}, {Key: "vote", Value: 1}},
			}},
		}},
	}

	var tests = []struct {
		name            string
		userID          int64
		voteChange      int
		mockResponse    []bson.D
		wantErr         string
		expectedScore   int
		expectedPercent int
	}{
		{
			name:       "Проверка на успешный upvote для userID 3",
			userID:     3,
			voteChange: 1,
			mockResponse: []bson.D{
				post,
				{
					{Key: "ok", Value: 1},
					{Key: "acknowledged", Value: true},
					{Key: "n", Value: 1},
					{Key: "nModified", Value: 1},
				},
			},
			wantErr:         "",
			expectedScore:   3,
			expectedPercent: 100,
		},
		{
			name:       "Проверка на успешный downvote для userID 1",
			userID:     1,
			voteChange: -1,
			mockResponse: []bson.D{
				post,
				{
					{Key: "ok", Value: 1},
					{Key: "acknowledged", Value: true},
					{Key: "n", Value: 1},
					{Key: "nModified", Value: 1},
				},
			},
			wantErr:         "",
			expectedScore:   -1,
			expectedPercent: 33,
		},
		{
			name:       "Проверка на успешный upvote для нового юзера",
			userID:     4,
			voteChange: 1,
			mockResponse: []bson.D{
				post,
				{
					{Key: "ok", Value: 1},
					{Key: "acknowledged", Value: true},
					{Key: "n", Value: 1},
					{Key: "nModified", Value: 1},
				},
			},
			wantErr:         "",
			expectedScore:   2,
			expectedPercent: 75,
		},
		{
			name:       "Проверка, что post.VoteCount < 0 обрабатывается правильно",
			userID:     1,
			voteChange: 1,
			mockResponse: []bson.D{
				badPost,
				{
					{Key: "ok", Value: 1},
					{Key: "acknowledged", Value: true},
					{Key: "n", Value: 1},
					{Key: "nModified", Value: 1},
				},
			},
			wantErr:         "",
			expectedScore:   1,
			expectedPercent: 0,
		},
		{
			name:       "Проверка на обработку ошибки при запросе на получение поста",
			userID:     1,
			voteChange: -1,
			mockResponse: []bson.D{
				{{Key: "ok", Value: 0}},
			},
			wantErr: `{"message": "post not found"}`,
		},
		{
			name:       "Проверка на обработку ошибки при обновлении поста",
			userID:     1,
			voteChange: -1,
			mockResponse: []bson.D{
				post,
				{{Key: "ok", Value: 0}},
			},
			wantErr: `{"message": "bad request"}`,
		},
		{
			name:       "Проверка на обработку ошибки ModifiedCount == 0",
			userID:     1,
			voteChange: -1,
			mockResponse: []bson.D{
				post,
				{
					{Key: "ok", Value: 1},
					{Key: "acknowledged", Value: true},
				},
			},
			wantErr: `{"message": "failed to update field"}`,
		},
	}

	for _, tc := range tests {
		mt.Run(tc.name, func(mt *mtest.T) {
			repo := NewMongoRepo(mt.Coll)
			mt.AddMockResponses(tc.mockResponse...)

			result, err := repo.VotePost(postID, tc.userID, tc.voteChange)
			if tc.wantErr != "" {
				assert.Error(t, err)
				assert.EqualError(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tc.expectedScore, result.Score, "expected score %d, got %d", tc.expectedScore, result.Score)
				assert.Equal(t, tc.expectedPercent, result.UpvotePercentage, "expected upvote percentage %d, got %d", tc.expectedPercent, result.UpvotePercentage)
			}
		})
	}
}

func TestUnVotePost(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	postID := primitive.NewObjectID()
	post := bson.D{
		{Key: "ok", Value: 1},
		{Key: "value", Value: bson.D{
			{Key: "_id", Value: postID},
			{Key: "Score", Value: 1},
			{Key: "Views", Value: 9},
			{Key: "Type", Value: "text"},
			{Key: "Title", Value: "Post 3"},
			{Key: "Category", Value: "fashion"},
			{Key: "Text", Value: "Text of the post 1"},
			{Key: "Created", Value: "2024-05-07T20:38:17Z"},
			{Key: "UpvotePercentage", Value: 100},
			{Key: "Upvotecount", Value: 2},
			{Key: "Votecount", Value: 3},
			{Key: "Author", Value: bson.M{
				"id":       3,
				"username": "ivan",
			}},
			{Key: "Votes", Value: bson.A{
				bson.D{{Key: "userID", Value: 1}, {Key: "vote", Value: 1}},
				bson.D{{Key: "userID", Value: 2}, {Key: "vote", Value: 1}},
				bson.D{{Key: "userID", Value: 3}, {Key: "vote", Value: -1}},
			}},
		}},
	}
	badPost := bson.D{
		{Key: "ok", Value: 1},
		{Key: "value", Value: bson.D{
			{Key: "_id", Value: postID},
			{Key: "Score", Value: 1},
			{Key: "Views", Value: 3},
			{Key: "Type", Value: "text"},
			{Key: "Title", Value: "Post 1"},
			{Key: "Category", Value: "fashion"},
			{Key: "Text", Value: "Text of the post 1"},
			{Key: "Created", Value: "2024-05-07T20:38:17Z"},
			{Key: "UpvotePercentage", Value: 0},
			{Key: "Upvotecount", Value: 0},
			{Key: "Votecount", Value: 0},
			{Key: "Author", Value: bson.M{
				"id":       1,
				"username": "ivan",
			}},
			{Key: "Votes", Value: bson.A{
				bson.D{{Key: "userID", Value: 1}, {Key: "vote", Value: 1}},
			}},
		}},
	}

	var tests = []struct {
		name               string
		userID             int64
		mockResponse       []bson.D
		wantErr            string
		expectedScore      int
		expectedPercent    int
		expectedVotesCount int
	}{
		{
			name:   "Проверка на успешный unvote для userID 3",
			userID: 3,
			mockResponse: []bson.D{
				post,
				{
					{Key: "ok", Value: 1},
					{Key: "acknowledged", Value: true},
					{Key: "n", Value: 1},
					{Key: "nModified", Value: 1},
				},
			},
			wantErr:            "",
			expectedScore:      2,
			expectedPercent:    100,
			expectedVotesCount: 2,
		},
		{
			name:   "Проверка, что post.VoteCount < 0 обрабатывается правильно",
			userID: 1,
			mockResponse: []bson.D{
				badPost,
				{
					{Key: "ok", Value: 1},
					{Key: "acknowledged", Value: true},
					{Key: "n", Value: 1},
					{Key: "nModified", Value: 1},
				},
			},
			wantErr:            "",
			expectedScore:      0,
			expectedPercent:    0,
			expectedVotesCount: 2,
		},
		{
			name:   "Проверка на обработку ошибки при запросе на получение поста",
			userID: 1,
			mockResponse: []bson.D{
				{{Key: "ok", Value: 0}},
			},
			wantErr: `{"message": "post not found"}`,
		},
		{
			name:   "Проверка на обработку ошибки при удалении комментария",
			userID: 1,
			mockResponse: []bson.D{
				post,
				{{Key: "ok", Value: 0}},
			},
			wantErr: `{"message": "failed to update field"}`,
		},
		{
			name:   "Проверка на обработку ошибки ModifiedCount == 0",
			userID: 1,
			mockResponse: []bson.D{
				post,
				{
					{Key: "ok", Value: 1},
					{Key: "acknowledged", Value: true},
				},
			},
			wantErr: `{"message": "failed to update field"}`,
		},
	}

	for _, tc := range tests {
		mt.Run(tc.name, func(mt *mtest.T) {
			repo := NewMongoRepo(mt.Coll)
			mt.AddMockResponses(tc.mockResponse...)

			result, err := repo.UnVotePost(postID, tc.userID)
			if tc.wantErr != "" {
				assert.Error(t, err)
				assert.EqualError(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tc.expectedScore, result.Score, "expected score %d, got %d", tc.expectedScore, result.Score)
				assert.Equal(t, tc.expectedPercent, result.UpvotePercentage, "expected upvote percentage %d, got %d", tc.expectedPercent, result.UpvotePercentage)
			}
		})
	}
}

func TestMakePost(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	newTextPost := PostForm{
		Type:     "text",
		Title:    "Post 1",
		Category: "fashion",
		Text:     "Text of the post 1",
	}

	newURLPost := PostForm{
		Type:     "link",
		Title:    "Post 1",
		Category: "fashion",
		URL:      "https://www.youtube.com/",
	}

	var tests = []struct {
		name         string
		userID       int64
		userName     string
		mockResponse []bson.D
		wantErr      string
		newPostData  PostForm
	}{
		{
			name:     "Проверка на успешне создание поста c текстом юзером с UserID 3",
			userID:   3,
			userName: "ivan",
			mockResponse: []bson.D{
				mtest.CreateSuccessResponse(),
			},
			wantErr:     "",
			newPostData: newTextPost,
		},
		{
			name:     "Проверка на успешне создание поста с url юзером с UserID 3",
			userID:   3,
			userName: "ivan",
			mockResponse: []bson.D{
				mtest.CreateSuccessResponse(),
			},
			wantErr:     "",
			newPostData: newURLPost,
		},
		{
			name:    "Проверка на обработку ошибки, когда данные для поста пустые",
			userID:  3,
			wantErr: `{"message": "bad request"}`,
		},
		{
			name:     "Проверка на обработку ошибки при запросе на создание поста",
			userID:   3,
			userName: "ivan",
			mockResponse: []bson.D{
				{{Key: "ok", Value: 0}},
			},
			newPostData: newTextPost,
			wantErr:     `{"message": "bad request"}`,
		},
	}

	for _, tc := range tests {
		mt.Run(tc.name, func(mt *mtest.T) {
			repo := NewMongoRepo(mt.Coll)
			mt.AddMockResponses(tc.mockResponse...)

			result, err := repo.MakePost(&tc.newPostData, tc.userName, tc.userID)
			if tc.wantErr != "" {
				assert.Error(t, err)
				assert.EqualError(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tc.userID, result.Author.ID, "expected userID %d, got %d", tc.userID, result.Author.ID)
				assert.Equal(t, tc.newPostData.Title, result.Title, "expected title %s, got %s", tc.newPostData.Title, result.Title)
			}
		})
	}
}

func TestDeletePost(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	var tests = []struct {
		name         string
		mockResponse []bson.D
		wantErr      string
	}{
		{
			name: "Проверка на успешное удаление поста",
			mockResponse: []bson.D{
				{
					{Key: "ok", Value: 1},
					{Key: "acknowledged", Value: true},
					{Key: "n", Value: 1},
				},
			},
			wantErr: "",
		},
		{
			name: "Проверка на обработку ошибки, что пост не удалился в бд",
			mockResponse: []bson.D{
				{
					{Key: "ok", Value: 0},
				},
			},
			wantErr: `{"message": "bad request"}`,
		},
		{
			name: "Проверка на обработку ошибки DeletedCount == 0",
			mockResponse: []bson.D{
				{
					{Key: "ok", Value: 1},
					{Key: "acknowledged", Value: true},
					{Key: "n", Value: 0},
				},
			},
			wantErr: `{"message": "failed to delete"}`,
		},
	}

	for _, tc := range tests {
		mt.Run(tc.name, func(mt *mtest.T) {
			repo := NewMongoRepo(mt.Coll)
			mt.AddMockResponses(tc.mockResponse...)

			result, err := repo.DeletePost(primitive.NewObjectID(), 1)
			if tc.wantErr != "" {
				assert.Error(t, err)
				assert.EqualError(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, true, result, "expected to be %v, got %v", true, result)
			}
		})
	}
}

func TestMakeComment(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	postID := primitive.NewObjectID()
	commentID := primitive.NewObjectID()
	post := bson.D{
		{Key: "ok", Value: 1},
		{Key: "value", Value: bson.D{
			{Key: "_id", Value: postID},
			{Key: "Score", Value: 2},
			{Key: "Views", Value: 9},
			{Key: "Type", Value: "text"},
			{Key: "Title", Value: "Post 3"},
			{Key: "Category", Value: "fashion"},
			{Key: "Text", Value: "Text of the post 1"},
			{Key: "Created", Value: "2024-05-07T20:38:17Z"},
			{Key: "UpvotePercentage", Value: 100},
			{Key: "Author", Value: bson.M{
				"id":       3,
				"username": "ivan",
			}},
			{Key: "Comments", Value: bson.A{
				bson.D{
					{Key: "_id", Value: commentID},
					{Key: "Body", Value: "some comment"},
					{Key: "Author", Value: bson.M{
						"id":       3,
						"username": "ivan",
					}},
				},
			}},
		}},
	}

	var tests = []struct {
		name          string
		userID        int64
		userName      string
		comment       string
		mockResponse  []bson.D
		wantErr       string
		commentsCount int
	}{
		{
			name:     "Проверка на успешное создание комментария",
			userID:   3,
			userName: "ivan",
			comment:  "some comment",
			mockResponse: []bson.D{
				post,
			},
			wantErr:       "",
			commentsCount: 1,
		},
		{
			name:     "Проверка на обработку ошибок при запросе к бд",
			userID:   3,
			userName: "ivan",
			comment:  "some comment",
			mockResponse: []bson.D{
				{{Key: "ok", Value: 0}},
			},
			wantErr: `{"message": "bad request"}`,
		},
		{
			name:     "Проверка на обработку сломанного bson",
			userID:   3,
			userName: "ivan",
			comment:  "some comment",
			wantErr:  `{"message": "bad request"}`,
			mockResponse: []bson.D{{
				{Key: "ok", Value: 1},
				{Key: "value", Value: bson.D{
					{Key: "_id", Value: "не ObjectId, а строка"},
					{Key: "UpvotePercentage", Value: "должно быть число, передано как строка"},
				}},
			},
			},
		},
	}

	for _, tc := range tests {
		mt.Run(tc.name, func(mt *mtest.T) {
			repo := NewMongoRepo(mt.Coll)
			mt.AddMockResponses(tc.mockResponse...)

			result, err := repo.MakeComment(postID, tc.comment, tc.userName, tc.userID)
			if tc.wantErr != "" {
				assert.Error(t, err)
				assert.EqualError(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tc.commentsCount, len(result.Comments), "expected %d comment, got %d", tc.commentsCount, len(result.Comments))
			}
		})
	}
}

func TestDeleteComment(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	var tests = []struct {
		name          string
		userID        int64
		mockResponse  []bson.D
		wantErr       string
		commentsCount int
	}{
		{
			name:   "Проверка на успешное удаление коммента",
			userID: 3,
			mockResponse: []bson.D{
				{
					{Key: "ok", Value: 1},
					{Key: "value", Value: bson.D{
						{Key: "_id", Value: primitive.NewObjectID()},
						{Key: "Score", Value: 2},
						{Key: "Views", Value: 9},
						{Key: "Type", Value: "text"},
						{Key: "Title", Value: "Post 3"},
						{Key: "Category", Value: "fashion"},
						{Key: "Text", Value: "Text of the post 1"},
						{Key: "Created", Value: "2024-05-07T20:38:17Z"},
						{Key: "UpvotePercentage", Value: 100},
						{Key: "Author", Value: bson.M{
							"id":       3,
							"username": "ivan",
						}},
						{Key: "Comments", Value: bson.A{}},
					}},
				},
			},
			wantErr:       "",
			commentsCount: 0,
		},
		{
			name:   "Проверка на обработку ошибки, что коммент не удалился в бд",
			userID: 3,
			mockResponse: []bson.D{
				{
					{Key: "ok", Value: 0},
				},
			},
			wantErr: `{"message": "failed to update field"}`,
		},
		{
			name:   "Проверка на обработку сломанного bson",
			userID: 3,
			mockResponse: []bson.D{
				{
					{Key: "ok", Value: 1},
					{Key: "value", Value: bson.D{
						{Key: "_id", Value: "не ObjectId, а строка"},
						{Key: "UpvotePercentage", Value: "должно быть число, передано как строка"},
					}},
				},
			},
			wantErr: `{"message": "bad request"}`,
		},
	}

	for _, tc := range tests {
		mt.Run(tc.name, func(mt *mtest.T) {
			repo := NewMongoRepo(mt.Coll)
			mt.AddMockResponses(tc.mockResponse...)

			result, err := repo.DeleteComment(primitive.NewObjectID(), primitive.NewObjectID(), tc.userID)
			if tc.wantErr != "" {
				assert.Error(t, err)
				assert.EqualError(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tc.commentsCount, len(result.Comments), "expected %d comment, got %d", tc.commentsCount, len(result.Comments))
			}
		})
	}
}
