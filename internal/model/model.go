package model

type (
	// JSONRequest структура необходимая для декодирования JSON.
	JSONRequest struct {
		URL string `json:"url"`
	}

	// JSONResponse структура необходимая для кодирования JSON.
	JSONResponse struct {
		Result string `json:"result"`
	}

	// JSONStructForFile структура необходимая для декодирования JSON в файл.
	JSONStructForFile struct {
		UUID        string `json:"uuid"`
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
		DeletedFlag bool   `db:"is_deleted"`
	}

	// ShotenBatchRequest структура необходимая для декодирования JSON для списка URL.
	ShotenBatchRequest struct {
		CorrelationID string `json:"correlation_id"`
		OriginalURL   string `json:"original_url"`
	}

	// ShortenResponseItem структура необходимая для кодирования JSON для списка URL.
	ShortenResponseItem struct {
		CorrelationID string `json:"correlation_id"`
		ShortURL      string `json:"short_url"`
	}

	// ShortenURLs структура необходимая для получения необходимого URL.
	ShortenURLs struct {
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}
)
