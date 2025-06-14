package rest

import (
	"context"
	"net/http"
	"os"
	"strconv"

	pb "statisticservice/proto"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

var grpcClient pb.StatisticsServiceClient

func InitGRPCClient() {
	conn, err := grpc.Dial(os.Getenv("GRPC_SERVER_ADDR"), grpc.WithInsecure())
	if err != nil {
		panic("failed to connect to gRPC server: " + err.Error())
	}
	grpcClient = pb.NewStatisticsServiceClient(conn)
}

func RegisterRoutes(r *gin.Engine) {
	stats := r.Group("/stats")
	{
		stats.GET("/post/:id", GetPostStats)
		stats.GET("/post/:id/views", GetPostViewsDynamics)
		stats.GET("/post/:id/likes", GetPostLikesDynamics)
		stats.GET("/post/:id/comments", GetPostCommentsDynamics)
		stats.GET("/top/posts", GetTopPosts)
		stats.GET("/top/users", GetTopUsers)
	}
}

func GetPostStats(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	resp, err := grpcClient.GetPostStats(context.Background(), &pb.GetPostStatsRequest{
		PostId: uint32(id),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func GetPostViewsDynamics(c *gin.Context) {
	getPostDynamics(c, grpcClient.GetPostViewsDynamics)
}

func GetPostLikesDynamics(c *gin.Context) {
	getPostDynamics(c, grpcClient.GetPostLikesDynamics)
}

func GetPostCommentsDynamics(c *gin.Context) {
	getPostDynamics(c, grpcClient.GetPostCommentsDynamics)
}

func getPostDynamics(c *gin.Context, fn func(ctx context.Context, in *pb.GetPostDynamicsRequest, opts ...grpc.CallOption) (*pb.GetDynamicsResponse, error)) {
	id, _ := strconv.Atoi(c.Param("id"))
	resp, err := fn(context.Background(), &pb.GetPostDynamicsRequest{
		PostId: uint32(id),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func GetTopPosts(c *gin.Context) {
	statType := parseStatType(c.Query("by"))
	resp, err := grpcClient.GetTopPosts(context.Background(), &pb.GetTopRequest{By: statType})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func GetTopUsers(c *gin.Context) {
	statType := parseStatType(c.Query("by"))
	resp, err := grpcClient.GetTopUsers(context.Background(), &pb.GetTopRequest{By: statType})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func parseStatType(s string) pb.GetTopRequest_StatType {
	switch s {
	case "likes":
		return pb.GetTopRequest_LIKES
	case "comments":
		return pb.GetTopRequest_COMMENTS
	default:
		return pb.GetTopRequest_VIEWS
	}
}
