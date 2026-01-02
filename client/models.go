package simple_ai

import "google.golang.org/genai"

type EventActions struct {
	StateDelta    map[string]any   `json:"stateDelta"`
	ArtifactDelta map[string]int64 `json:"artifactDelta"`
}

type Event struct {
	ID                 string                   `json:"id"`
	Time               int64                    `json:"time"`
	InvocationID       string                   `json:"invocationId"`
	Branch             string                   `json:"branch"`
	Author             string                   `json:"author"`
	Partial            bool                     `json:"partial"`
	LongRunningToolIDs []string                 `json:"longRunningToolIds"`
	Content            *genai.Content           `json:"content"`
	GroundingMetadata  *genai.GroundingMetadata `json:"groundingMetadata"`
	TurnComplete       bool                     `json:"turnComplete"`
	Interrupted        bool                     `json:"interrupted"`
	ErrorCode          string                   `json:"errorCode"`
	ErrorMessage       string                   `json:"errorMessage"`
	Actions            EventActions             `json:"actions"`
}
type Session struct {
	ID        string         `json:"id"`
	AppName   string         `json:"appName"`
	UserID    string         `json:"userId"`
	UpdatedAt int64          `json:"lastUpdateTime"`
	Events    []Event        `json:"events"`
	State     map[string]any `json:"state"`
}

type AdkRequest struct {
	AppName string `json:"appName"`

	UserId string `json:"userId"`

	SessionId string `json:"sessionId"`

	NewMessage genai.Content `json:"newMessage"`

	Streaming bool `json:"streaming,omitempty"`

	StateDelta *map[string]any `json:"stateDelta,omitempty"`
}
