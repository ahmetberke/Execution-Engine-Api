package models

type CommandRequest struct {
	UserID  string `json:"user_id"`
	Command string `json:"command"`
}

type CommandResponse struct {
	UserID string `json:"user_id"`
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
}
