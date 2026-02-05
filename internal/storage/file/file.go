package file

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/aga-absolut/url-cutter/internal/config"
	"github.com/aga-absolut/url-cutter/internal/model"
	"go.uber.org/zap"
)

type File struct {
	UUID   int
	config *config.Config
	logger zap.SugaredLogger
}

func NewFile(config *config.Config, logger zap.SugaredLogger) *File {
	return &File{config: config, logger: logger}
}

func (f *File) checkFile(originalURL string) (string, error) {
	file, err := os.Open(f.config.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		var data model.JSONStructForFile
		if err := json.Unmarshal([]byte(line), &data); err != nil {
			f.logger.Errorw("Error parcing line:", "Error", err)
		}
		if data.OriginalURL == originalURL {
			f.logger.Infow("Succes work originalURL:", "URL", originalURL)
			return data.ShortURL, fmt.Errorf("not unique URL")
		}
	}
	if err := scanner.Err(); err != nil {
		f.logger.Errorw("Error parcing file:", "Error", err)
	}
	return "", nil
}

func (f *File) Set(ctx context.Context, shortURL, originalURL string, userID int) (string, error) {
	if shortKey, err := f.checkFile(originalURL); err != nil {
		if errors.Is(err, nil) {
			return shortKey, os.ErrExist
		}
		return "", err
	}

	f.UUID++
	UUID := strconv.Itoa(f.UUID)
	data := model.JSONStructForFile{
		UUID:        UUID,
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}

	result, err := json.Marshal(data)
	if err != nil {
		f.logger.Errorw("Error marshal file:", "Error", shortURL)
		return "", err
	}
	result = append(result, '\n')

	file, err := os.OpenFile(f.config.FilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		f.logger.Errorw("Error open file:", "Error", shortURL)
		return "", err
	}
	defer file.Close()

	if _, err = file.Write(result); err != nil {
		f.logger.Errorw("Error write file:", "Error", shortURL)
		return "", err
	}
	return "", nil
}

func (f *File) Get(ctx context.Context, shortURL string) (string, bool) {
	file, err := os.Open(f.config.FilePath)
	if err != nil {
		f.logger.Errorw("Error open file:", "Error", err)
		return "", false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var data model.JSONStructForFile
		if err := json.Unmarshal([]byte(line), &data); err != nil {
			f.logger.Errorw("Error parcing line:", "Error", err)
			continue
		}

		if data.ShortURL == shortURL {
			f.logger.Infow("Succes work shorURL:", "URL", shortURL)
			return data.OriginalURL, true
		}

	}
	if err := scanner.Err(); err != nil {
		f.logger.Errorw("Error parcing file:", "Error", err)
	}
	return "", false
}
