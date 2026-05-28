package exa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// ContentResult holds the fetched content for a single URL.
type ContentResult struct {
	URL     string `json:"url"`
	Content string `json:"content"`
}

type contentsRequest struct {
	IDs  []string `json:"ids"`
	Text bool     `json:"text,omitempty"`
}

type contentsResult struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type contentsResponse struct {
	Results []contentsResult `json:"results"`
}

func exaPost(body contentsRequest) (*contentsResponse, error) {
	apiKey := os.Getenv("EXA_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("EXA_API_KEY environment variable is required")
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, "https://api.exa.ai/contents", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling Exa API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Exa API returned status %d", resp.StatusCode)
	}

	var result contentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}

// FetchContents retrieves text content for multiple URLs in a single Exa API call.
// Reads EXA_API_KEY from environment variables.
func FetchContents(urls []string) ([]ContentResult, error) {
	result, err := exaPost(contentsRequest{IDs: urls, Text: true})
	if err != nil {
		return nil, err
	}
	out := make([]ContentResult, len(result.Results))
	for i, r := range result.Results {
		out[i] = ContentResult{URL: r.ID, Content: r.Text}
	}
	return out, nil
}

// FetchContent retrieves the text content of a single URL using the Exa Contents API.
func FetchContent(pageURL string) (string, error) {
	results, err := FetchContents([]string{pageURL})
	if err != nil {
		return "", err
	}
	if len(results) == 0 {
		return "", fmt.Errorf("no content returned for URL: %s", pageURL)
	}
	return results[0].Content, nil
}
