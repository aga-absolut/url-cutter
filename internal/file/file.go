package file

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"strconv"

	"github.com/aga-absolut/url-cutter/internal/config"
)

type URLRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	config      config.Config
}

func NewURLRecord(cfg *config.Config) *URLRecord {
	return &URLRecord{config: *cfg}
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

func (f *URLRecord) Save(shortURL, originalURL string) error {
	counter, err := UpdateCounter()
	if err != nil {
		return err
	}
	data := URLRecord{
		UUID:        counter,
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}
	result, err := json.Marshal(data)
	if err != nil {
		return err
	}
	result = append(result, '\n')
	file, err := os.OpenFile(f.config.FilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err = file.Write(result); err != nil {
		return err
	}
	return nil
}

func (f *URLRecord) ReadFile(shortURL string) (string, bool) {
	file, err := os.Open(f.config.FilePath)
	if err != nil {
		return "", false
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var data URLRecord
		if err := json.Unmarshal([]byte(line), &data); err != nil {
			log.Printf("Error parcing line: %v", err)
			continue
		}

		if data.ShortURL == shortURL {
			return data.OriginalURL, true
		}

	}
	if err := scanner.Err(); err != nil {
		log.Printf("Error parcing file: %v", err)
	}
	return "", false
}
