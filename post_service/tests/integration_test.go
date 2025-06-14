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
