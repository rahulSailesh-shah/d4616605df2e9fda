package tools

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func FetchWikipediaWord(title string, position int) (string, error) {
	url := fmt.Sprintf("https://en.wikipedia.org/api/rest_v1/page/summary/%s", strings.ReplaceAll(title, " ", "_"))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("request %q: %w", title, err)
	}
	req.Header.Set("User-Agent", "NeonHealthAgent/1.0 (https://github.com/rahulSailesh-shah/neon_health)")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch %q: %w", title, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("wikipedia returned %d for %q", resp.StatusCode, title)
	}

	var summary struct {
		Extract string `json:"extract"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&summary); err != nil {
		return "", fmt.Errorf("decode %q: %w", title, err)
	}

	words := strings.Fields(summary.Extract)
	if position < 1 || position > len(words) {
		return "", fmt.Errorf("position %d out of range (1-%d) for %q", position, len(words), title)
	}

	word := words[position-1]
	word = strings.TrimRight(word, ".,;:!?\"'()[]")
	return word, nil
}
