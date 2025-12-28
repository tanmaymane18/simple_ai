package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	"google.golang.org/genai"
)

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

type AdkClient struct {
	sv_url      string
	adk_session Session
}

func InitAdkClient() AdkClient {
	return AdkClient{
		sv_url: "http://localhost:8080/api",
		adk_session: Session{
			ID:      uuid.NewString(),
			AppName: "CodePipelineAgent",
			UserID:  "simple_user",
		},
	}
}

func (ac *AdkClient) CreateSession() bool {
	url := fmt.Sprintf("%s/apps/%s/users/%s/sessions", ac.sv_url, ac.adk_session.AppName, ac.adk_session.UserID)
	if resp, err := http.Post(url, "application/json", nil); err != nil {
		return false
	} else {
		defer resp.Body.Close()
		var adk_session Session
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error while reading response")
			return false
		}
		err = json.Unmarshal(body, &adk_session)
		if err != nil {
			fmt.Printf("Error while Unmarshal: %v\n", err)
			return false
		}
		ac.adk_session = adk_session
		return true
	}
}

func (ac *AdkClient) PrepareAdkRequest(message string) AdkRequest {
	return AdkRequest{
		AppName:   ac.adk_session.AppName,
		UserId:    ac.adk_session.UserID,
		SessionId: ac.adk_session.ID,
		Streaming: false,
		NewMessage: genai.Content{
			Parts: []*genai.Part{{Text: message}},
			Role:  "user",
		},
	}

}

func (ac *AdkClient) MakeRequest(message string) {
	url := fmt.Sprintf("%s/run", ac.sv_url)
	adk_request := ac.PrepareAdkRequest(message)
	adk_request_bytes, err := json.Marshal(adk_request)
	if err != nil {
		fmt.Printf("Error while marshal: %v", err)
		return
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(adk_request_bytes))
	if err != nil {
		fmt.Printf("Error while post: %v", err)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	fmt.Println(string(body))
	if err != nil {
		fmt.Println("Error while reading response")
		return
	}
	var events []Event
	if err := json.Unmarshal(body, &events); err != nil {
		fmt.Printf("Error while Unmarshal response: %v", err)
	} else {
		fmt.Printf("Reponse: \n\n%v", events[len(events)-1].Content.Parts[0].Text)
	}
}

func main() {
	ac := InitAdkClient()
	ac.CreateSession()
	ac.MakeRequest("Add colors in py_lazygit")
}
