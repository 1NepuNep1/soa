package handlers

import (
	"context"
	"errors"
	"log"
	"postservice/database"
	"postservice/kafka"
	"postservice/models"
	pb "postservice/proto"

	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

type Server struct {
	pb.UnimplementedPostServiceServer
}

func (s *Server) CreatePost(ctx context.Context, req *pb.CreatePostRequest) (*pb.PostResponse, error) {
	post := models.Post{
		Title:       req.Title,
		Description: req.Description,
		CreatorID:   req.CreatorId,
		IsPrivate:   req.IsPrivate,
		Tags:        req.Tags,
	}

	if err := database.DB.Create(&post).Error; err != nil {
		log.Println("Failed to create post:", err)
		return nil, err
	}

	go kafka.SendClientRegistered(req.CreatorId)

	return &pb.PostResponse{Post: convertModelToProto(post)}, nil
}

func (s *Server) GetPostByID(ctx context.Context, req *pb.GetPostByIDRequest) (*pb.PostResponse, error) {
	var post models.Post
	err := database.DB.First(&post, req.Id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("post not found")
	} else if err != nil {
		return nil, err
	}

	// Проверка приватности
	if post.IsPrivate && post.CreatorID != req.RequesterId {
		return nil, errors.New("permission denied")
	}

	go kafka.SendInteractionEvent("post_viewed", req.RequesterId, req.Id)

	return &pb.PostResponse{Post: convertModelToProto(post)}, nil
}

func (s *Server) UpdatePost(ctx context.Context, req *pb.UpdatePostRequest) (*pb.PostResponse, error) {
	var post models.Post
	err := database.DB.First(&post, req.Id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("post not found")
	} else if err != nil {
		return nil, err
	}

	if post.CreatorID != req.RequesterId {
		return nil, errors.New("permission denied")
	}

	post.Title = req.Title
	post.Description = req.Description
	post.IsPrivate = req.IsPrivate
	post.Tags = req.Tags

	if err := database.DB.Save(&post).Error; err != nil {
		return nil, err
	}

	return &pb.PostResponse{Post: convertModelToProto(post)}, nil
}

func (s *Server) DeletePost(ctx context.Context, req *pb.DeletePostRequest) (*pb.DeletePostResponse, error) {
	var post models.Post
	err := database.DB.First(&post, req.Id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &pb.DeletePostResponse{Success: false}, errors.New("post not found")
	} else if err != nil {
		return &pb.DeletePostResponse{Success: false}, err
	}

	if post.CreatorID != req.RequesterId {
		return &pb.DeletePostResponse{Success: false}, errors.New("permission denied")
	}

	if err := database.DB.Delete(&post).Error; err != nil {
		return &pb.DeletePostResponse{Success: false}, err
	}

	return &pb.DeletePostResponse{Success: true}, nil
}

func (s *Server) ListPosts(ctx context.Context, req *pb.ListPostsRequest) (*pb.ListPostsResponse, error) {
	var posts []models.Post
	var total int64

	query := database.DB.Model(&models.Post{}).Where(
		database.DB.Where("is_private = ?", false).Or("creator_id = ?", req.RequesterId),
	)

	query.Count(&total)

	err := query.Offset(int((req.Page - 1) * req.PageSize)).
		Limit(int(req.PageSize)).
		Order("created_at DESC").
		Find(&posts).Error

	if err != nil {
		return nil, err
	}

	return &pb.ListPostsResponse{
		Posts: convertModelsToProtos(posts),
		Total: uint32(total),
	}, nil
}

func convertModelToProto(post models.Post) *pb.Post {
	return &pb.Post{
		Id:          uint32(post.ID),
		Title:       post.Title,
		Description: post.Description,
		CreatorId:   uint32(post.CreatorID),
		IsPrivate:   post.IsPrivate,
		Tags:        post.Tags,
		CreatedAt:   timestamppb.New(post.CreatedAt),
		UpdatedAt:   timestamppb.New(post.UpdatedAt),
	}
}

func convertModelsToProtos(posts []models.Post) []*pb.Post {
	protos := make([]*pb.Post, len(posts))
	for i, p := range posts {
		protos[i] = convertModelToProto(p)
	}
	return protos
}

func (s *Server) LikePost(ctx context.Context, req *pb.LikePostRequest) (*pb.LikePostResponse, error) {
	var post models.Post
	err := database.DB.First(&post, req.PostId).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &pb.LikePostResponse{Success: false}, errors.New("post not found")
	} else if err != nil {
		return &pb.LikePostResponse{Success: false}, err
	}

	// тут типа лайк короч

	go kafka.SendInteractionEvent("post_liked", req.ClientId, req.PostId)

	return &pb.LikePostResponse{Success: true}, nil
}

func (s *Server) CommentPost(ctx context.Context, req *pb.CommentPostRequest) (*pb.CommentPostResponse, error) {
	var post models.Post
	if err := database.DB.First(&post, req.PostId).Error; err != nil {
		return &pb.CommentPostResponse{Success: false}, errors.New("post not found")
	}

	comment := models.Comment{
		PostID:   req.PostId,
		ClientID: req.ClientId,
		Content:  req.Content,
	}

	if err := database.DB.Create(&comment).Error; err != nil {
		return &pb.CommentPostResponse{Success: false}, err
	}

	go kafka.SendInteractionEvent("post_commented", req.ClientId, req.PostId)

	return &pb.CommentPostResponse{Success: true}, nil
}

func (s *Server) ListComments(ctx context.Context, req *pb.ListCommentsRequest) (*pb.ListCommentsResponse, error) {
	var comments []models.Comment
	var total int64

	database.DB.Model(&models.Comment{}).
		Where("post_id = ?", req.PostId).
		Count(&total)

	err := database.DB.Where("post_id = ?", req.PostId).
		Order("created_at DESC").
		Offset(int((req.Page - 1) * req.PageSize)).
		Limit(int(req.PageSize)).
		Find(&comments).Error

	if err != nil {
		return nil, err
	}

	pbComments := make([]*pb.Comment, len(comments))
	for i, c := range comments {
		pbComments[i] = &pb.Comment{
			Id:        uint32(c.ID),
			PostId:    c.PostID,
			ClientId:  c.ClientID,
			Content:   c.Content,
			CreatedAt: timestamppb.New(c.CreatedAt),
		}
	}

	return &pb.ListCommentsResponse{
		Comments: pbComments,
		Total:    uint32(total),
	}, nil
}
