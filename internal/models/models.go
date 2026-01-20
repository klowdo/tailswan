package models

type ConnectionRequest struct {
	Name string `json:"name"`
}

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

type ConnectionsResponse struct {
	Success     bool                     `json:"success"`
	Connections []map[string]interface{} `json:"connections"`
}

type SAsResponse struct {
	Success bool                     `json:"success"`
	SAs     []map[string]interface{} `json:"sas"`
}
