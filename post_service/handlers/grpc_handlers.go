package handlers

import (
	"context"
	"errors"
	"log"
	"postservice/database"
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
