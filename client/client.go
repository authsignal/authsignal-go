package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// todo do you want the Errors remapped to a Authsignal one or is the strategy to leave dependencies escalating their own errors.
// todo only forward body elements required along, not all requests.

const RequestTimeout = 10 * time.Second

type Client struct {
	apiKey      string
	apiUrl      string
	redirectUrl string
}

func New(apiUrl, apiKey string, redirectUrl string) Client {
	return Client{apiKey: apiKey, apiUrl: apiUrl, redirectUrl: redirectUrl}
}

func (c Client) defaultHeaders() http.Header {
	return http.Header{
		"Accept":       {"*/*"},
		"Content-Type": {"application/json"},
		"User-Agent":   {c.userAgent()},
	}
}

func (c Client) userAgent() string {
	return "Authsignal Go v1" // todo make module version dynamic
}

func (c Client) GetUser(userId string) (string, error) {
	return c.get(userId)
}

func (c Client) TrackAction(request TrackRequest) (TrackResponse, error) {
	body, err := json.Marshal(request)
	if err != nil {
		return TrackResponse{}, err
	}

	path := fmt.Sprintf("%s/actions/%s", request.UserId, request.Action)
	response, err2 := c.post(path, bytes.NewBuffer(body))
	if err2 != nil {
		return TrackResponse{}, err2
	}

	var data TrackResponse
	err3 := json.Unmarshal([]byte(response), &data)
	if err3 != nil {
		return TrackResponse{}, err3
	}

	return data, nil
}

func (c Client) GetAction(request GetActionRequest) (GetActionResponse, error) {
	path := fmt.Sprintf("%s/actions/%s/%s", request.UserId, request.Action, request.IdempotencyKey)
	response, err := c.get(path)
	if err != nil {
		return GetActionResponse{}, err
	}

	var data GetActionResponse
	err2 := json.Unmarshal([]byte(response), &data)
	if err2 != nil {
		return GetActionResponse{}, err2
	}

	return data, nil
}

func (c Client) EnrollVerifiedAuthenticator(request EnrollVerifiedAuthenticatorRequest) (EnrollVerifiedAuthenticatorResponse, error) {
	body, err := json.Marshal(request)
	if err != nil {
		return EnrollVerifiedAuthenticatorResponse{}, err
	}

	path := fmt.Sprintf("%s/authenticators", request.UserId)
	response, err2 := c.post(path, bytes.NewBuffer(body))
	if err2 != nil {
		return EnrollVerifiedAuthenticatorResponse{}, err
	}

	var data EnrollVerifiedAuthenticatorResponse
	err3 := json.Unmarshal([]byte(response), &data)
	if err3 != nil {
		return EnrollVerifiedAuthenticatorResponse{}, err3
	}

	return data, nil
}

func (c Client) LoginWithEmail(request LoginWithEmailRequest) (LoginWithEmailResponse, error) {
	body, err := json.Marshal(request)
	if err != nil {
		return LoginWithEmailResponse{}, err
	}

	response, err2 := c.post(fmt.Sprintf("/email/%s/challenge", request.Email), bytes.NewBuffer(body))
	if err2 != nil {
		return LoginWithEmailResponse{}, err
	}

	var data LoginWithEmailResponse
	err3 := json.Unmarshal([]byte(response), &data)
	if err3 != nil {
		return LoginWithEmailResponse{}, err3
	}

	return data, err
}

func (c Client) ValidateChallenge(request ValidateChallengeRequest) string {
	return "NOT YET IMPLEMENTED"
}

func (c Client) get(path string) (string, error) {
	return c.makeRequest("GET", path, nil)
}

func (c Client) post(path string, body io.Reader) (string, error) {
	return c.makeRequest("POST", path, body)
}

func (c Client) makeRequest(method, path string, body io.Reader) (string, error) {
	client := http.Client{}
	client.Timeout = RequestTimeout
	req, err := http.NewRequest(method, fmt.Sprintf("%s/v1/users/%s", c.apiUrl, path), body)
	if err != nil {
		return "", err
	}

	req.Header = c.defaultHeaders()
	req.SetBasicAuth(c.apiKey, "")
	// todo Context for HTTP requests to put the Timeout + other config in.

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			return //todo handle error or pass up.
		}
	}()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Todo deal with HTTP status code.
	// todo return byte[] or unpack into a response object generically.
	sb := string(responseBody)
	return sb, nil
}
