package model

type JSONRequest struct {
	URL string `json:"url"`
}
type JSONResponse struct {
	Result string `json:"result"`
}

// может лучще куда-то в storage их убрать? или еще куда-то
