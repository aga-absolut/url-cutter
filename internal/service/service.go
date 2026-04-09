package service

import (
	"context"
	"errors"
	"net"
	"strings"

	"github.com/aga-absolut/url-cutter/internal/config"
	"github.com/aga-absolut/url-cutter/internal/errs"
	"github.com/aga-absolut/url-cutter/internal/model"
	"github.com/aga-absolut/url-cutter/internal/repository"
	"github.com/aga-absolut/url-cutter/internal/util"
	"github.com/aga-absolut/url-cutter/middleware/jwt"
)

// Структура обработчика
type Service struct {
	Config     *config.Config
	Storage    repository.Storage
	deleteChan chan string
}

// NewHandler создает новую структуру Handler
func NewService(config *config.Config, storage repository.Storage, deleteChan chan string) *Service {
	service := &Service{
		Config:     config,
		Storage:    storage,
		deleteChan: deleteChan,
	}
	return service
}

func (s *Service) Ping() error {
	if err := s.Storage.Ping(); err != nil {
		return err
	}
	return nil
}

func (s *Service) GetURL(ctx context.Context, shortURL string) (string, bool) {
	resURL, exist := s.Storage.Get(ctx, shortURL)
	return resURL, exist
}

func (s *Service) DeleteURLs(arrShortURLs []string) {
	for _, shortURL := range arrShortURLs {
		s.deleteChan <- shortURL
	}
}

func (s *Service) GetUserURLs(ctx context.Context, cookie string) ([]model.ShortenURLs, error) {
	userID, err := jwt.GetUserID(cookie)
	if err != nil {
		return nil, errs.ErrInGettingUserID
	}

	URLs, err := s.Storage.GetByUserID(ctx, userID)
	if err != nil {
		return nil, errs.ErrInGettingURLs
	}

	return URLs, nil
}

func (s *Service) SetURL(ctx context.Context, originalURL []byte, cookie string) (string, error) {
	if len(originalURL) == 0 {
		return "", errs.ErrEmptyBody
	}

	userID, err := jwt.GetUserID(cookie)
	if err != nil {
		return "", errs.ErrInGettingUserID
	}

	shortURL := util.Generate(string(originalURL))
	if shortKey, err := s.Storage.Set(ctx, shortURL, string(originalURL), userID); err != nil {
		if errors.Is(err, errs.ErrURLAlreadyExists) {
			return s.Config.Host + "/" + shortKey, err
		}
		return "", err
	}

	return s.Config.Host + "/" + shortURL, nil
}

func (s *Service) SetBathcURLs(ctx context.Context, batch []model.ShotenBatchRequest, cookie string) ([]model.ShortenResponseItem, error) {
	userID, err := jwt.GetUserID(cookie)
	if err != nil {
		return nil, errs.ErrInGettingUserID
	}

	if len(batch) == 0 {
		return nil, errs.ErrEmptyBatch
	}

	response, err := s.Storage.SetBatchURL(ctx, batch, userID)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *Service) SetURLFromJSON(ctx context.Context, req model.JSONRequest, cookie string) (model.JSONResponse, error) {
	if len(req.URL) == 0 {
		return model.JSONResponse{}, errs.ErrEmptyBody
	}

	userID, err := jwt.GetUserID(cookie)
	if err != nil {
		return model.JSONResponse{}, errs.ErrInGettingUserID
	}

	shortURL := util.Generate(req.URL)
	if shortKey, err := s.Storage.Set(ctx, shortURL, req.URL, userID); err != nil {
		if errors.Is(err, errs.ErrURLAlreadyExists) {
			return model.JSONResponse{Result: s.Config.Host + "/" + shortKey}, err
		}
		return model.JSONResponse{}, err
	}

	return model.JSONResponse{Result: s.Config.Host + "/" + shortURL}, nil
}

func (s *Service) GetStats(ctx context.Context, ipStr, forwarded string) (model.ResponseStats, error) {
	if s.Config.TrustedSubnet == "" {
		return model.ResponseStats{}, errs.ErrTrustedSubnetIsEmpty
	}

	_, ipNet, err := net.ParseCIDR(s.Config.TrustedSubnet)
	if err != nil {
		return model.ResponseStats{}, err
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		ipStrs := strings.Split(forwarded, ",")
		if len(ipStrs) > 0 {
			ip = net.ParseIP(ipStrs[0])
		}
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
