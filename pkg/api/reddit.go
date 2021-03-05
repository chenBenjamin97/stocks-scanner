package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type AccessToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

//* Notice: request to generate access token you should use 'https://www.reddit..'
//* Once you have generated an access token, you should use 'https://oauth.reddit..'
//* reddit api documentation: 'App-only OAuth token requests never receive a refresh_token' - we will generate a new access token when expired

func GenerateAccessToken(username, userPassword, appID, appSecret string) (*AccessToken, error) {
	client := http.Client{}

	grantAccessBody := fmt.Sprintf("grant_type=password&username=%s&password=%s", username, userPassword)
	req, err := http.NewRequest("POST", "https://www.reddit.com/api/v1/access_token", strings.NewReader(grantAccessBody))
	if err != nil {
		return nil, fmt.Errorf("GenerateAccessToken: Error, got '%v'", err)
	}
	req.SetBasicAuth(appID, appSecret)
	req.Header.Add("User-Agent", "Script") //needed header in order to avoid bad response, can change the value

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GenerateAccessToken: Error, got '%v'", err)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("GenerateAccessToken: Error, got '%v'", err)
	}

	var token AccessToken
	if err := json.Unmarshal(respBodyBytes, &token); err != nil {
		return nil, fmt.Errorf("GenerateAccessToken: Error, got '%v'", err)
	}

	return &token, nil
}

func TrendingSubreddits(accessToken string) (string, error) {
	client := http.Client{}

	req, err := http.NewRequest("GET", "https://oauth.reddit.com/r/wallstreetbets/new?limit=10", nil)
	if err != nil {
		return "", fmt.Errorf("TrendingSubreddits: Error, got '%v'", err)
	}
	req.Header.Add("User-Agent", "Script") //needed header in order to avoid bad response, can change the value
	req.Header.Add("Authorization", fmt.Sprintf("bearer %s", accessToken))

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("TrendingSubreddits: Error, got '%v'", err)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("TrendingSubreddits: Error, got '%v'", err)
	}

	f, err := os.OpenFile("test.json", os.O_CREATE|os.O_TRUNC, 0622)
	if err != nil {
		return "", fmt.Errorf("TrendingSubreddits: Error, got '%v'", err)
	}

	f.WriteString(string(respBodyBytes))
	f.Sync()
	f.Close()

	return string(respBodyBytes), nil
}
