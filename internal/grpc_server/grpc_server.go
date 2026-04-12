package grpcserver

import (
	"context"
	"net"

	"github.com/aga-absolut/url-cutter/internal/service"
	pb "github.com/aga-absolut/url-cutter/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type URLCutterServer struct {
	pb.UnimplementedURLCutterServer

	service *service.Service
}

func NewURLCutterServer(service *service.Service) pb.URLCutterServer {
	return &URLCutterServer{service: service}
}

func (u *URLCutterServer) GetHandler(ctx context.Context, in *pb.GetRequest) (*pb.GetResponse, error) {
	url, ok := u.service.GetURL(ctx, in.ShortUrl)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "not found")
	}
	return &pb.GetResponse{OriginalUrl: url}, nil
}

func (u *URLCutterServer) GetUserURLs(ctx context.Context, in *pb.GetUserURLSRequest) (*pb.GetUserURLSResponse, error) {
	urls, err := u.service.GetUserURLs(ctx, int(in.UserId))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "not found")
	}

	var list []*pb.ShortenURLs
	for _, i := range urls {
		list = append(list, &pb.ShortenURLs{
			ShortUrl:    i.ShortURL,
			OriginalUrl: i.OriginalURL,
		})
	}

	response := pb.GetUserURLSResponse{
		OriginalUrls: list,
	}

	return &response, nil
}

func (u *URLCutterServer) GetStatsHandler(ctx context.Context, in *pb.GetStatsRequest) (*pb.GetStatsResponse, error) {
	var ip net.IP
	if peerInfo, ok := peer.FromContext(ctx); ok {
		if tcpAddr, ok := peerInfo.Addr.(*net.TCPAddr); ok {
			ip = tcpAddr.IP
		}
	}

	stats, err := u.service.GetStats(ctx, ip)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "not found")
	}

	return &pb.GetStatsResponse{TotalUsers: int64(stats.Users), TotalUrls: int64(stats.URLs)}, nil
}

func (u *URLCutterServer) PostHandler(ctx context.Context, in *pb.PostRequest) (*pb.PostResponse, error) {
	shortURL, err := u.service.SetURL(ctx, []byte(in.OriginalUrl), int(in.UserId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	return &pb.PostResponse{ShortUrl: shortURL}, nil
}

func (u *URLCutterServer) PostBatchHandler(ctx context.Context, in *pb.PostBatchRequest) (*pb.PostBatchResponse, error) {
	var response []*pb.BatchResponseItem

	for _, i := range in.Items {
		shortURL, err := u.service.SetURL(ctx, []byte(i.OriginalUrl), int(i.UserID))
		if err != nil {
			return nil, status.Errorf(codes.Internal, "internal server error")
		}

		response = append(response, &pb.BatchResponseItem{
			UserID:   i.UserID,
			ShortUrl: shortURL,
		})
	}

	return &pb.PostBatchResponse{Items: response}, nil
}

func (u *URLCutterServer) DeleteUserURLs(ctx context.Context, in *pb.DeleteUserURLsRequest) (*pb.DeleteUserURLsResponse, error) {
	u.service.DeleteURLs(in.OriginalUrls)
	return &pb.DeleteUserURLsResponse{}, nil
}

func (u *URLCutterServer) CheckConnecToDB(ctx context.Context, in *emptypb.Empty) (*pb.CheckConnecToDBResponse, error) {
	if err := u.service.Ping(); err != nil {
		return &pb.CheckConnecToDBResponse{}, status.Error(codes.Internal, "internal server error")
	}
	return &pb.CheckConnecToDBResponse{}, nil
}
