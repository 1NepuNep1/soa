package handlers

import (
	"context"
	"statisticservice/stats"

	pb "statisticservice/proto"
)

type Server struct {
	pb.UnimplementedStatisticsServiceServer
}

func (s *Server) GetPostStats(ctx context.Context, req *pb.GetPostStatsRequest) (*pb.GetPostStatsResponse, error) {
	metrics, err := stats.GetPostStats(ctx, req.PostId)
	if err != nil {
		return nil, err
	}
	return &pb.GetPostStatsResponse{
		Views:    uint32(metrics.Views),
		Likes:    uint32(metrics.Likes),
		Comments: uint32(metrics.Comments),
	}, nil
}

func (s *Server) GetPostViewsDynamics(ctx context.Context, req *pb.GetPostDynamicsRequest) (*pb.GetDynamicsResponse, error) {
	return getDynamics(ctx, req.PostId, 1) // 1 = view
}

func (s *Server) GetPostLikesDynamics(ctx context.Context, req *pb.GetPostDynamicsRequest) (*pb.GetDynamicsResponse, error) {
	return getDynamics(ctx, req.PostId, 2) // 2 = like
}

func (s *Server) GetPostCommentsDynamics(ctx context.Context, req *pb.GetPostDynamicsRequest) (*pb.GetDynamicsResponse, error) {
	return getDynamics(ctx, req.PostId, 3) // 3 = comment
}

func getDynamics(ctx context.Context, postID uint32, eventType int8) (*pb.GetDynamicsResponse, error) {
	dayCounts, err := stats.GetPostDynamics(ctx, postID, eventType)
	if err != nil {
		return nil, err
	}
	resp := &pb.GetDynamicsResponse{}
	for _, dc := range dayCounts {
		resp.Data = append(resp.Data, &pb.DayCount{
			Date:  dc.Date,
			Count: uint32(dc.Count),
		})
	}
	return resp, nil
}

func (s *Server) GetTopPosts(ctx context.Context, req *pb.GetTopRequest) (*pb.GetTopPostsResponse, error) {
	eventType := mapStatTypeToEventType(req.By)
	topPosts, err := stats.GetTopPosts(ctx, eventType, 10)
	if err != nil {
		return nil, err
	}

	resp := &pb.GetTopPostsResponse{}
	for _, post := range topPosts {
		resp.Posts = append(resp.Posts, &pb.PostStat{
			PostId: post.PostID,
			Count:  uint32(post.Count),
		})
	}
	return resp, nil
}

func (s *Server) GetTopUsers(ctx context.Context, req *pb.GetTopRequest) (*pb.GetTopUsersResponse, error) {
	eventType := mapStatTypeToEventType(req.By)
	topUsers, err := stats.GetTopUsers(ctx, eventType, 10)
	if err != nil {
		return nil, err
	}

	resp := &pb.GetTopUsersResponse{}
	for _, user := range topUsers {
		resp.Users = append(resp.Users, &pb.UserStat{
			UserId: user.UserID,
			Count:  uint32(user.Count),
		})
	}
	return resp, nil
}

func mapStatTypeToEventType(statType pb.GetTopRequest_StatType) int8 {
	switch statType {
	case pb.GetTopRequest_VIEWS:
		return 1
	case pb.GetTopRequest_LIKES:
		return 2
	case pb.GetTopRequest_COMMENTS:
		return 3
	default:
		return 0
	}
}
