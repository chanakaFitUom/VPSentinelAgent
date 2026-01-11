package models

// Command represents a command sent from the backend to the agent
type Command struct {
	Type    string                 `json:"type"`    // "stop", "restart", "update_config", "ping"
	ID      string                 `json:"id"`      // Command ID for tracking
	Payload map[string]interface{} `json:"payload"` // Command-specific payload
}

// CommandResponse represents the agent's response to a command
type CommandResponse struct {
	CommandID string `json:"command_id"`
	Status    string `json:"status"` // "success", "error", "processing"
	Message   string `json:"message,omitempty"`
}
