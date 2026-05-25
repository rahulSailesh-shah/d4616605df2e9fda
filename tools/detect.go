package tools

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	recallRe = regexp.MustCompile(`(?i)(?:earlier you transmitted|you previously transmitted|earlier.*?transmitted).*?crew member's\s+(.+?)\.\s+(?:Speak|Transmit)\s+the\s+(\d+)(?:st|nd|rd|th)\s+word`)
)

type FastResult struct {
	Type  string
	Value string
}

func (r *FastResult) JSON() string {
	if r.Type == "enter_digits" {
		return fmt.Sprintf(`{"type":"enter_digits","digits":"%s"}`, r.Value)
	}
	return fmt.Sprintf(`{"type":"speak_text","text":"%s"}`, r.Value)
}

type ResponseTracker struct {
	entries []trackedResponse
}

type trackedResponse struct {
	Category string
	Text     string
}

var categoryKeywords = []struct {
	keyword  string
	category string
}{
	{"education", "education"},
	{"skills", "skills"},
	{"work experience", "work experience"},
	{"best project", "best project"},
	{"project", "project"},
	{"reason", "reason"},
	{"granted access", "reason"},
	{"background", "background"},
}

func (t *ResponseTracker) DetectAndStore(neonMsg, responseText string) {
	msgLower := strings.ToLower(neonMsg)
	category := "unknown"
	for _, kw := range categoryKeywords {
		if strings.Contains(msgLower, kw.keyword) {
			category = kw.category
			break
		}
	}
	t.entries = append(t.entries, trackedResponse{Category: category, Text: responseText})
}

func (t *ResponseTracker) TryRecall(message string) *FastResult {
	m := recallRe.FindStringSubmatch(message)
	if m == nil {
		return nil
	}

	categoryHint := strings.ToLower(strings.TrimSpace(m[1]))
	wordPos, err := strconv.Atoi(m[2])
	if err != nil || wordPos < 1 {
		return nil
	}

	var matchedText string
	for _, entry := range t.entries {
		if strings.Contains(categoryHint, entry.Category) || strings.Contains(entry.Category, categoryHint) {
			matchedText = entry.Text
			break
		}
	}

	if matchedText == "" {
		return nil
	}

	words := strings.Fields(matchedText)
	if wordPos > len(words) {
		return nil
	}

	word := words[wordPos-1]
	word = strings.TrimRight(word, ".,;:!?\"'()[]")
	return &FastResult{Type: "speak_text", Value: word}
}
