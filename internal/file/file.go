package file

import (
	"bufio"
	"encoding/json"
	"os"
	"strconv"

	"github.com/aga-absolut/url-cutter/internal/config"
	"go.uber.org/zap"
)

type (
	JSONStructForFile struct {
		UUID        string `json:"uuid"`
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}
	File struct {
		UUID   int
		config *config.Config
		logger zap.SugaredLogger
	}
)

func NewFile(config *config.Config, logger zap.SugaredLogger) *File {
	return &File{config: config, logger: logger}
}

func UpdateCounter() (string, error) {
	file, err := os.ReadFile("counter.txt")
	if err != nil {
		os.WriteFile("counter.txt", []byte("1"), 0666)
		return "1", nil
	}
	militaryCounter, err := strconv.Atoi(string(file))
	if err != nil {
		return "", err
	}
	militaryCounter += 1
	counter := strconv.Itoa(militaryCounter)
	err = os.WriteFile("counter.txt", []byte(counter), 0666)
	if err != nil {
		return "", err
	}
	return counter, nil
}

func (f *File) Save(shortURL, originalURL string) error {
	f.UUID++
	UUID := strconv.Itoa(f.UUID)

	data := JSONStructForFile{
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
	if _, err = file.Write(result); err != nil {
		f.logger.Errorw("Error write file:", "Error", shortURL)
		return err
	}
	return nil
}

func (f *File) ReadFile(shortURL string) (string, bool) {
	file, err := os.Open(f.config.FilePath)
	if err != nil {
		f.logger.Errorw("Error open file:", "Error", shortURL)
		return "", false
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var data JSONStructForFile
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
