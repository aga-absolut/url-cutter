package service

import (
	"context"
	"errors"
	"net"

	"github.com/aga-absolut/url-cutter/internal/config"
	"github.com/aga-absolut/url-cutter/internal/errs"
	"github.com/aga-absolut/url-cutter/internal/generate"
	"github.com/aga-absolut/url-cutter/internal/model"
	"github.com/aga-absolut/url-cutter/internal/repository"
	"github.com/aga-absolut/url-cutter/internal/transport/http/middleware/jwt"
)

type Service interface {
	Ping() error
	DeleteURLs(arrShortURLs []string)
	GetURL(ctx context.Context, shortURL string) (string, bool)
	GetStats(ctx context.Context, ip net.IP) (model.ResponseStats, error)
	GetUserURLs(ctx context.Context, userID int) ([]model.ShortenURLs, error)
	SetURL(ctx context.Context, originalURL []byte, userID int) (string, error)
	SetURLFromJSON(ctx context.Context, req model.JSONRequest, userID int) (model.JSONResponse, error)
	SetBatchURLs(ctx context.Context, batch []model.ShortenBatchRequest, userID int) ([]model.ShortenResponseItem, error)
}

type service struct {
	Config     *config.Config
	Storage    repository.Storage
	deleteChan chan string
}

func NewService(config *config.Config, storage repository.Storage, deleteChan chan string) *service {
	service := &service{
		Config:     config,
		Storage:    storage,
		deleteChan: deleteChan,
	}
	return service
}

func (s *service) Ping() error {
	if err := s.Storage.Ping(); err != nil {
		return err
	}
	return nil
}

func (s *service) GetURL(ctx context.Context, shortURL string) (string, bool) {
	resURL, exist := s.Storage.Get(ctx, shortURL)
	return resURL, exist
}

func (s *service) DeleteURLs(arrShortURLs []string) {
	for _, shortURL := range arrShortURLs {
		s.deleteChan <- shortURL
	}
}

func (s *service) GetUserURLs(ctx context.Context, userID int) ([]model.ShortenURLs, error) {
	URLs, err := s.Storage.GetByUserID(ctx, userID)
	if err != nil {
		return nil, errs.ErrInGettingURLs
	}

	return URLs, nil
}

func (s *service) SetURL(ctx context.Context, originalURL []byte, userID int) (string, error) {
	if len(originalURL) == 0 {
		return "", errs.ErrEmptyBody
	}

	shortURL := generate.Generate(string(originalURL))
	if shortKey, err := s.Storage.Set(ctx, shortURL, string(originalURL), userID); err != nil {
		if errors.Is(err, errs.ErrURLAlreadyExists) {
			return s.Config.Host + "/" + shortKey, err
		}
		return "", err
	}

	return s.Config.Host + "/" + shortURL, nil
}

func (s *service) SetBatchURLs(ctx context.Context, batch []model.ShortenBatchRequest, userID int) ([]model.ShortenResponseItem, error) {
	if len(batch) == 0 {
		return nil, errs.ErrEmptyBatch
	}

	response, err := s.Storage.SetBatchURL(ctx, batch, userID)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *service) SetURLFromJSON(ctx context.Context, req model.JSONRequest, userID int) (model.JSONResponse, error) {
	if len(req.URL) == 0 {
		return model.JSONResponse{}, errs.ErrEmptyBody
	}

	shortURL := generate.Generate(req.URL)
	if shortKey, err := s.Storage.Set(ctx, shortURL, req.URL, userID); err != nil {
		if errors.Is(err, errs.ErrURLAlreadyExists) {
			return model.JSONResponse{Result: s.Config.Host + "/" + shortKey}, err
		}
		return model.JSONResponse{}, err
	}

	return model.JSONResponse{Result: s.Config.Host + "/" + shortURL}, nil
}

func (s *service) GetStats(ctx context.Context, ip net.IP) (model.ResponseStats, error) {
	if s.Config.TrustedSubnet == "" {
		return model.ResponseStats{}, errs.ErrTrustedSubnetIsEmpty
	}

	_, ipNet, err := net.ParseCIDR(s.Config.TrustedSubnet)
	if err != nil {
		return model.ResponseStats{}, err
	}

	if ip == nil {
		return model.ResponseStats{}, errs.ErrNilIP
	}

	if !ipNet.Contains(ip) {
		return model.ResponseStats{}, errs.ErrIPNotAllowed
	}

	urls, err := s.Storage.GetURLsCount(ctx)
	if err != nil {
		return model.ResponseStats{}, errs.ErrInGettingUrlsCount
	}

	return model.ResponseStats{URLs: urls, Users: jwt.UserID}, nil
}
