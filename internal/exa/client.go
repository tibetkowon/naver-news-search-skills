package exa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type highlightsOption struct {
	MaxCharacters int    `json:"maxCharacters,omitempty"`
	Query         string `json:"query,omitempty"`
}

type contentsRequest struct {
	IDs        []string          `json:"ids"`
	Text       bool              `json:"text,omitempty"`
	Highlights *highlightsOption `json:"highlights,omitempty"`
}

type contentsResult struct {
	ID              string    `json:"id"`
	Text            string    `json:"text"`
	Highlights      []string  `json:"highlights"`
	HighlightScores []float64 `json:"highlightScores"`
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

// FetchContent retrieves the text content of a URL using the Exa Contents API.
// It reads EXA_API_KEY from environment variables.
func FetchContent(pageURL string) (string, error) {
	result, err := exaPost(contentsRequest{IDs: []string{pageURL}, Text: true})
	if err != nil {
		return "", err
	}
	if len(result.Results) == 0 {
		return "", fmt.Errorf("no content returned for URL: %s", pageURL)
	}
	return result.Results[0].Text, nil
}

// FetchHighlights retrieves key highlight snippets from the given URL.
// query guides which sentences to extract (empty string = general relevance).
// maxChars limits total characters per URL (0 uses default 500).
func FetchHighlights(pageURL, query string, maxChars int) ([]string, error) {
	if maxChars == 0 {
		maxChars = 500
	}
	opts := &highlightsOption{MaxCharacters: maxChars}
	if query != "" {
		opts.Query = query
	}
	result, err := exaPost(contentsRequest{IDs: []string{pageURL}, Highlights: opts})
	if err != nil {
		return nil, err
	}
	if len(result.Results) == 0 {
		return nil, fmt.Errorf("no highlights returned for URL: %s", pageURL)
	}
	return result.Results[0].Highlights, nil
}
