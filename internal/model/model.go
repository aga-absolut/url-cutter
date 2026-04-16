package model

type (
	// DataForCert струкура необходимая для декодирования данных из файла для сертификата
	DataForCert struct {
		CertFile string `json:"certFile"`
		KeyFile  string `json:"keyFile"`
	}

	// ResponseStats структура необходимая для кодирования JSON для данный о количестве urlов и userов.
	ResponseStats struct {
		URLs  int `json:"urls"`
		Users int `json:"users"`
	}
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
	ShortenBatchRequest struct {
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
