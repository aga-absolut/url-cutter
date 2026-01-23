package file

import (
	"bufio"
	"encoding/json"
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

func (f *File) checkFile(originalURL string) error {
	file, err := os.Open(f.config.FilePath)
	if err != nil {
		f.logger.Errorw("Error open file:", "Error", err)
		return err
	}
	
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
			return fmt.Errorf("Not unique URL.")
		}
	}
	if err := scanner.Err(); err != nil {
		f.logger.Errorw("Error parcing file:", "Error", err)
	}
	return nil
}

func (f *File) Set(shortURL, originalURL string) error {
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
		return err
	}
	result = append(result, '\n')

	file, err := os.OpenFile(f.config.FilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		f.logger.Errorw("Error open file:", "Error", shortURL)
		return err
	}
	defer file.Close()

	if err := f.checkFile(originalURL); err != nil {
		return fmt.Errorf("Not unique URL.")
	}

	if _, err = file.Write(result); err != nil {
		f.logger.Errorw("Error write file:", "Error", shortURL)
		return err
	}
	return nil
}

func (f *File) Get(shortURL string) (string, bool) {
	file, err := os.Open(f.config.FilePath)
	if err != nil {
		f.logger.Errorw("Error open file:", "Error", err)
		return "", false
	}
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
