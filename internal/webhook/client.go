package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Target string

const (
	TargetSlack   Target = "slack"
	TargetDiscord Target = "discord"
	TargetGeneric Target = "webhook"
)

func Send(target Target, webhookURL, text string) error {
	if strings.TrimSpace(webhookURL) == "" {
		return fmt.Errorf("webhook URL is required")
	}
	if strings.TrimSpace(text) == "" {
		return fmt.Errorf("message text is empty")
	}

	var payload any
	switch target {
	case TargetSlack:
		payload = map[string]string{"text": text}
	case TargetDiscord:
		payload = map[string]string{"content": text}
	case TargetGeneric:
		payload = map[string]string{"text": text}
	default:
		return fmt.Errorf("unsupported webhook target: %s", target)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling webhook payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating webhook request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "naver-news-cli/3.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("calling webhook: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	return nil
}
