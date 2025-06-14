package tests

import (
	"context"
	"testing"

	"postservice/database"
	"postservice/handlers"
	"postservice/models"
	pb "postservice/proto"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var grpcHandler pb.PostServiceServer

func setupTestDB() {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to sqlite: " + err.Error())
	}
	database.DB = db
	err = database.DB.AutoMigrate(&models.Post{})
	if err != nil {
		panic("failed to migrate: " + err.Error())
	}
	err = database.DB.AutoMigrate(&models.Comment{})
	if err != nil {
		panic("failed to migrate: " + err.Error())
	}
}

func setup() {
	setupTestDB()
	grpcHandler = &handlers.Server{}
}

func TestCreateAndGetPost(t *testing.T) {
	setup()

	createResp, err := grpcHandler.CreatePost(context.Background(), &pb.CreatePostRequest{
		Title:       "Integration Title",
		Description: "From test",
		CreatorId:   99,
		IsPrivate:   false,
	})
	assert.NoError(t, err)

	getResp, err := grpcHandler.GetPostByID(context.Background(), &pb.GetPostByIDRequest{
		Id:          createResp.Post.Id,
		RequesterId: 99,
	})
	assert.NoError(t, err)
	assert.Equal(t, "Integration Title", getResp.Post.Title)
}

func TestPermissionDeniedOnPrivatePost(t *testing.T) {
	setup()

	resp, err := grpcHandler.CreatePost(context.Background(), &pb.CreatePostRequest{
		Title:       "Private post",
		Description: "shhh",
		CreatorId:   1,
		IsPrivate:   true,
	})
	assert.NoError(t, err)

	_, err = grpcHandler.GetPostByID(context.Background(), &pb.GetPostByIDRequest{
		Id:          resp.Post.Id,
		RequesterId: 2,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
}

func TestUpdatePost(t *testing.T) {
	setup()

	created, _ := grpcHandler.CreatePost(context.Background(), &pb.CreatePostRequest{
		Title:       "Original",
		Description: "Old description",
		CreatorId:   7,
		IsPrivate:   false,
	})

	updated, err := grpcHandler.UpdatePost(context.Background(), &pb.UpdatePostRequest{
		Id:          created.Post.Id,
		Title:       "Updated Title",
		Description: "Updated Desc",
		IsPrivate:   false,
		RequesterId: 7,
	})
	assert.NoError(t, err)
	assert.Equal(t, "Updated Title", updated.Post.Title)
	assert.Equal(t, "Updated Desc", updated.Post.Description)
}

func TestDeletePost(t *testing.T) {
	setup()

	created, _ := grpcHandler.CreatePost(context.Background(), &pb.CreatePostRequest{
		Title:       "To be deleted",
		Description: "bye",
		CreatorId:   13,
	})

	delResp, err := grpcHandler.DeletePost(context.Background(), &pb.DeletePostRequest{
		Id:          created.Post.Id,
		RequesterId: 13,
	})
	assert.NoError(t, err)
	assert.True(t, delResp.Success)

	_, err = grpcHandler.GetPostByID(context.Background(), &pb.GetPostByIDRequest{
		Id:          created.Post.Id,
		RequesterId: 13,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestListPosts(t *testing.T) {
	setup()

	for i := 0; i < 3; i++ {
		_, _ = grpcHandler.CreatePost(context.Background(), &pb.CreatePostRequest{
			Title:       "Public " + string(rune(i)),
			Description: "Visible",
			CreatorId:   1,
			IsPrivate:   false,
		})
	}
	_, _ = grpcHandler.CreatePost(context.Background(), &pb.CreatePostRequest{
		Title:       "Secret",
		Description: "You shall not pass",
		CreatorId:   2,
		IsPrivate:   true,
	})

	resp, err := grpcHandler.ListPosts(context.Background(), &pb.ListPostsRequest{
		Page:        1,
		PageSize:    10,
		RequesterId: 1,
	})
	assert.NoError(t, err)
	assert.Equal(t, uint32(3), resp.Total)
	assert.Len(t, resp.Posts, 3)
}

func TestLikePost(t *testing.T) {
	setup()

	created, _ := grpcHandler.CreatePost(context.Background(), &pb.CreatePostRequest{
		Title:       "Like me",
		Description: "like test",
		CreatorId:   101,
	})

	resp, err := grpcHandler.LikePost(context.Background(), &pb.LikePostRequest{
		PostId:   created.Post.Id,
		ClientId: 555,
	})

	assert.NoError(t, err)
	assert.True(t, resp.Success)
}

func TestViewTriggersAccess(t *testing.T) {
	setup()

	created, _ := grpcHandler.CreatePost(context.Background(), &pb.CreatePostRequest{
		Title:       "View me",
		Description: "visible to creator",
		CreatorId:   42,
		IsPrivate:   true,
	})

	resp, err := grpcHandler.GetPostByID(context.Background(), &pb.GetPostByIDRequest{
		Id:          created.Post.Id,
		RequesterId: 42,
	})
	assert.NoError(t, err)
	assert.Equal(t, "View me", resp.Post.Title)
}

func TestCommentAndListComments(t *testing.T) {
	setup()

	created, _ := grpcHandler.CreatePost(context.Background(), &pb.CreatePostRequest{
		Title:       "Discussion",
		Description: "open for comments",
		CreatorId:   7,
	})

	commentResp, err := grpcHandler.CommentPost(context.Background(), &pb.CommentPostRequest{
		PostId:   created.Post.Id,
		ClientId: 99,
		Content:  "Great post!",
	})
	assert.NoError(t, err)
	assert.True(t, commentResp.Success)

	listResp, err := grpcHandler.ListComments(context.Background(), &pb.ListCommentsRequest{
		PostId:   created.Post.Id,
		Page:     1,
		PageSize: 10,
	})
	assert.NoError(t, err)
	assert.Equal(t, uint32(1), listResp.Total)
	assert.Len(t, listResp.Comments, 1)
	assert.Equal(t, "Great post!", listResp.Comments[0].Content)
	assert.Equal(t, uint32(99), listResp.Comments[0].ClientId)
}
