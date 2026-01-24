package models

type ConnectionRequest struct {
	Name string `json:"name"`
}

type Response struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
	Success bool   `json:"success"`
}

type ConnectionsResponse struct {
	Connections []map[string]interface{} `json:"connections"`
	Success     bool                     `json:"success"`
}

type SAsResponse struct {
	SAs     []map[string]interface{} `json:"sas"`
	Success bool                     `json:"success"`
}

type SSEMessage struct {
	Event string
	Data  []byte
}
