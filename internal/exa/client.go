package exa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type contentsRequest struct {
	IDs  []string `json:"ids"`
	Text bool     `json:"text"`
}

type contentsResult struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type contentsResponse struct {
	Results []contentsResult `json:"results"`
}

// FetchContent retrieves the text content of a URL using the Exa Contents API.
// It reads EXA_API_KEY from environment variables.
func FetchContent(pageURL string) (string, error) {
	apiKey := os.Getenv("EXA_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("EXA_API_KEY environment variable is required")
	}

	body := contentsRequest{
		IDs:  []string{pageURL},
		Text: true,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, "https://api.exa.ai/contents", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("calling Exa API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Exa API returned status %d", resp.StatusCode)
	}

	var result contentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}

	if len(result.Results) == 0 {
		return "", fmt.Errorf("no content returned for URL: %s", pageURL)
	}

	return result.Results[0].Text, nil
}
