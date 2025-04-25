package rest

import (
	"context"
	"net/http"
	"os"
	"strconv"

	pb "postservice/proto"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

var grpcClient pb.PostServiceClient

func InitGRPCClient() {
	conn, err := grpc.Dial(os.Getenv("GRPC_SERVER_ADDR"), grpc.WithInsecure())
	if err != nil {
		panic("failed to connect gRPC server: " + err.Error())
	}
	grpcClient = pb.NewPostServiceClient(conn)
}

func RegisterRoutes(r *gin.Engine) {
	r.POST("/posts", CreatePost)
	r.GET("/posts/:id", GetPostByID)
	r.PUT("/posts/:id", UpdatePost)
	r.DELETE("/posts/:id", DeletePost)
	r.GET("/posts", ListPosts)
	r.POST("/posts/:id/like", LikePost)
	r.POST("/posts/:id/comments", CommentPost)
	r.GET("/posts/:id/comments", ListComments)
}

func CreatePost(c *gin.Context) {
	var req pb.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := grpcClient.CreatePost(context.Background(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp.Post)
}

func GetPostByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	requesterId, _ := strconv.Atoi(c.Query("requesterId"))

	resp, err := grpcClient.GetPostByID(context.Background(), &pb.GetPostByIDRequest{
		Id:          uint32(id),
		RequesterId: uint32(requesterId),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp.Post)
}

func UpdatePost(c *gin.Context) {
	var req pb.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, _ := strconv.Atoi(c.Param("id"))
	req.Id = uint32(id)

	resp, err := grpcClient.UpdatePost(context.Background(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp.Post)
}

func DeletePost(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	requesterId, _ := strconv.Atoi(c.Query("requesterId"))

	resp, err := grpcClient.DeletePost(context.Background(), &pb.DeletePostRequest{
		Id:          uint32(id),
		RequesterId: uint32(requesterId),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func ListPosts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	requesterId, _ := strconv.Atoi(c.Query("requesterId"))

	resp, err := grpcClient.ListPosts(context.Background(), &pb.ListPostsRequest{
		Page:        uint32(page),
		PageSize:    uint32(pageSize),
		RequesterId: uint32(requesterId),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func LikePost(c *gin.Context) {
	postId, _ := strconv.Atoi(c.Param("id"))
	clientId, _ := strconv.Atoi(c.Query("clientId"))

	resp, err := grpcClient.LikePost(context.Background(), &pb.LikePostRequest{
		PostId:   uint32(postId),
		ClientId: uint32(clientId),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": resp.Success})
}

func CommentPost(c *gin.Context) {
	postId, _ := strconv.Atoi(c.Param("id"))
	var body struct {
		ClientId uint32 `json:"clientId"`
		Content  string `json:"content"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := grpcClient.CommentPost(context.Background(), &pb.CommentPostRequest{
		PostId:   uint32(postId),
		ClientId: body.ClientId,
		Content:  body.Content,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": resp.Success})
}

func ListComments(c *gin.Context) {
	postId, _ := strconv.Atoi(c.Param("id"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	resp, err := grpcClient.ListComments(context.Background(), &pb.ListCommentsRequest{
		PostId:   uint32(postId),
		Page:     uint32(page),
		PageSize: uint32(pageSize),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
