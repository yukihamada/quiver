package api

type GenerateRequest struct {
	Prompt string `json:"prompt" binding:"required"`
	Model  string `json:"model" binding:"required"`
	Token  string `json:"token" binding:"required"`
}

type GenerateResponse struct {
	Completion string      `json:"completion"`
	Receipt    interface{} `json:"receipt"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
