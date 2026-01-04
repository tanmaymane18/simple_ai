package simple_ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"google.golang.org/genai"
)

type MsgStatus bool

type MsgSuccessResponse struct {
	Response string
}

type MsgFailureResponse struct {
	FailureMsg string
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

func (ac *AdkClient) CreateSession() tea.Msg {
	url := fmt.Sprintf("%s/apps/%s/users/%s/sessions", ac.sv_url, ac.adk_session.AppName, ac.adk_session.UserID)
	if resp, err := http.Post(url, "application/json", nil); err != nil {
		return MsgStatus(false)
	} else {
		defer resp.Body.Close()
		var adk_session Session
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return MsgStatus(false)
		}
		err = json.Unmarshal(body, &adk_session)
		if err != nil {
			return MsgStatus(false)
		}
		ac.adk_session = adk_session
		return MsgStatus(true)
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

func (ac *AdkClient) MakeRequest(message string) tea.Msg {
	url := fmt.Sprintf("%s/run", ac.sv_url)
	adk_request := ac.PrepareAdkRequest(message)
	adk_request_bytes, err := json.Marshal(adk_request)
	if err != nil {
		return MsgFailureResponse{
			FailureMsg: fmt.Sprintf("Error while marshaling: %v", err),
		}
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(adk_request_bytes))
	if err != nil {
		return MsgFailureResponse{
			FailureMsg: fmt.Sprintf("Error while making post request: %v", err),
		}
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return MsgFailureResponse{
			FailureMsg: fmt.Sprintf("Error while reading response body: %v", err),
		}
	}
	var events []Event
	if err := json.Unmarshal(body, &events); err != nil {
		return MsgFailureResponse{
			FailureMsg: fmt.Sprintf("Error while Unmarshal reponse: %v", err),
		}
	} else {
		return MsgSuccessResponse{
			Response: events[len(events)-1].Content.Parts[0].Text,
		}
	}
}
