package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type MockProvider struct{}

func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

func (p *MockProvider) Generate(ctx context.Context, input GenerateInput) (*GenerateResult, error) {
	if strings.Contains(input.UserPrompt, "Error Code:") {
		return generateMockErrorOptimization(input.UserPrompt)
	}

	payload := buildMockInvestigationPayload(input.UserPrompt)
	encoded, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, err
	}

	return &GenerateResult{
		Text: string(encoded),
	}, nil
}

func generateMockErrorOptimization(prompt string) (*GenerateResult, error) {
	// Extract the Error Code to customize the mock response
	code := "UNKNOWN_ERROR"
	if matches := regexp.MustCompile(`Error Code:\s+([^\n]+)`).FindStringSubmatch(prompt); len(matches) > 1 {
		code = strings.TrimSpace(matches[1])
	}

	mockMessage := "Something went wrong."
	if code == "AUTH_INVALID_CREDENTIALS" {
		mockMessage = "The email or password you entered is incorrect."
	} else if code == "AUTH_ACCOUNT_LOCKED" {
		mockMessage = "Your account has been temporarily locked due to multiple failed login attempts."
	}

	payload := map[string]interface{}{
		"message": mockMessage,
		"details": "This is a mock response from the AI Error Optimizer.",
		"suggestions": []map[string]interface{}{
			{
				"title":       "Reset your password",
				"description": "If you forgot your password, you can easily reset it.",
				"action":      "navigate",
				"url":         "/auth/forgot-password",
				"priority":    "primary",
			},
			{
				"title":       "Try again",
				"description": "Double check your spelling and try again.",
				"action":      "retry",
				"priority":    "secondary",
			},
		},
		"docs_url": "/docs/errors/" + code,
	}

	encoded, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, err
	}

	return &GenerateResult{
		Text: string(encoded),
	}, nil
}

type mockInvestigationPayload struct {
	Summary           string   `json:"summary"`
	Timeline          []string `json:"timeline"`
	SuspiciousSignals []string `json:"suspicious_signals"`
	Recommendations   []string `json:"recommendations"`
}

type mockPromptLog struct {
	Time        string
	Action      string
	Resource    string
	Status      string
	ActorUserID string
	IPAddress   string
	Description string
}

var (
	mockCountRegex    = regexp.MustCompile(`Audit log count:\s+(\d+)`)
	mockOverviewRegex = regexp.MustCompile(`Overview:\s+failed_logs=(\d+)\s+unique_ip_addresses=(\d+)\s+unique_actor_user_ids=(\d+)\s+first_seen=([^\s]+)\s+last_seen=([^\s]+)`)
)

func buildMockInvestigationPayload(prompt string) mockInvestigationPayload {
	logCount, failedCount, uniqueIPs, uniqueActors, firstSeen, lastSeen := parseMockOverview(prompt)
	logs := parseMockLogs(prompt)

	summary := fmt.Sprintf(
		"Mock investigation reviewed %d audit logs between %s and %s. %d events were marked as failed across %d IP addresses and %d actor IDs.",
		logCount,
		emptyMockFallback(firstSeen, "the selected time range"),
		emptyMockFallback(lastSeen, "the selected time range"),
		failedCount,
		uniqueIPs,
		uniqueActors,
	)
	if len(logs) > 0 {
		first := logs[0]
		summary = fmt.Sprintf(
			"Mock investigation reviewed %d %s %s events on resource %s between %s and %s, with %d failed events across %d IP addresses.",
			logCount,
			emptyMockFallback(first.Status, "recorded"),
			emptyMockFallback(first.Action, "activity"),
			emptyMockFallback(first.Resource, "n/a"),
			emptyMockFallback(firstSeen, "the selected time range"),
			emptyMockFallback(lastSeen, "the selected time range"),
			failedCount,
			uniqueIPs,
		)
	}

	timeline := buildMockTimeline(logs)
	signals := buildMockSignals(logs, failedCount, uniqueIPs)
	recommendations := buildMockRecommendations(signals)

	return mockInvestigationPayload{
		Summary:           summary,
		Timeline:          timeline,
		SuspiciousSignals: signals,
		Recommendations:   recommendations,
	}
}

func parseMockOverview(prompt string) (int, int, int, int, string, string) {
	logCount := parseMockInt(prompt, mockCountRegex, 1)
	matches := mockOverviewRegex.FindStringSubmatch(prompt)
	if len(matches) != 6 {
		return logCount, 0, 0, 0, "", ""
	}

	return logCount,
		parseMockAtoi(matches[1]),
		parseMockAtoi(matches[2]),
		parseMockAtoi(matches[3]),
		strings.TrimSpace(matches[4]),
		strings.TrimSpace(matches[5])
}

func parseMockLogs(prompt string) []mockPromptLog {
	lines := strings.Split(prompt, "\n")
	logs := make([]mockPromptLog, 0, 4)
	current := mockPromptLog{}
	inBlock := false

	flush := func() {
		if !inBlock {
			return
		}
		logs = append(logs, current)
		current = mockPromptLog{}
		inBlock = false
	}

	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if strings.HasPrefix(line, "- log ") {
			flush()
			inBlock = true
			continue
		}
		if strings.HasPrefix(line, "sample_event: ") {
			logs = append(logs, parseMockSampleEvent(strings.TrimPrefix(line, "sample_event: ")))
			continue
		}
		if !inBlock {
			continue
		}
		switch {
		case strings.HasPrefix(line, "time: "):
			current.Time = strings.TrimSpace(strings.TrimPrefix(line, "time: "))
		case strings.HasPrefix(line, "action: "):
			current.Action = strings.TrimSpace(strings.TrimPrefix(line, "action: "))
		case strings.HasPrefix(line, "resource: "):
			current.Resource = strings.TrimSpace(strings.TrimPrefix(line, "resource: "))
		case strings.HasPrefix(line, "status: "):
			current.Status = strings.TrimSpace(strings.TrimPrefix(line, "status: "))
		case strings.HasPrefix(line, "actor_user_id: "):
			current.ActorUserID = strings.TrimSpace(strings.TrimPrefix(line, "actor_user_id: "))
		case strings.HasPrefix(line, "ip_address: "):
			current.IPAddress = strings.TrimSpace(strings.TrimPrefix(line, "ip_address: "))
		case strings.HasPrefix(line, "description: "):
			current.Description = strings.Trim(strings.TrimSpace(strings.TrimPrefix(line, "description: ")), `"`)
		}
	}
	flush()

	if len(logs) > 3 {
		return logs[:3]
	}
	return logs
}

func parseMockSampleEvent(raw string) mockPromptLog {
	item := mockPromptLog{
		Time: raw,
	}
	parts := strings.SplitN(raw, " - ", 2)
	if len(parts) != 2 {
		return item
	}

	item.Time = strings.TrimSpace(parts[0])
	rest := parts[1]
	fields := strings.Fields(rest)
	if len(fields) >= 4 {
		item.Status = fields[0]
		item.Action = fields[1]
		if idx := strings.Index(rest, " on "); idx >= 0 {
			afterOn := rest[idx+4:]
			if end := strings.Index(afterOn, " from ip "); end >= 0 {
				item.Resource = strings.TrimSpace(afterOn[:end])
				afterIP := afterOn[end+9:]
				if actorIdx := strings.Index(afterIP, " (actor_user_id: "); actorIdx >= 0 {
					item.IPAddress = strings.TrimSpace(afterIP[:actorIdx])
					item.ActorUserID = strings.TrimSuffix(afterIP[actorIdx+18:], ")")
				} else {
					item.IPAddress = strings.TrimSpace(afterIP)
				}
			}
		}
	}
	return item
}

func buildMockTimeline(logs []mockPromptLog) []string {
	if len(logs) == 0 {
		return []string{
			"Reviewed the selected audit log window.",
			"Grouped the events into a short incident timeline for demo purposes.",
		}
	}

	lines := make([]string, 0, len(logs))
	for _, logEntry := range logs {
		lines = append(lines, fmt.Sprintf(
			"%s - %s %s on %s from ip %s (actor_user_id: %s)%s",
			emptyMockFallback(logEntry.Time, "unknown time"),
			emptyMockFallback(logEntry.Status, "recorded"),
			emptyMockFallback(logEntry.Action, "activity"),
			emptyMockFallback(logEntry.Resource, "n/a"),
			emptyMockFallback(logEntry.IPAddress, "n/a"),
			emptyMockFallback(logEntry.ActorUserID, "n/a"),
			mockDescriptionSuffix(logEntry.Description),
		))
	}
	return lines
}

func buildMockSignals(logs []mockPromptLog, failedCount int, uniqueIPs int) []string {
	signals := make([]string, 0, 3)
	if failedCount >= 2 {
		signals = append(signals, fmt.Sprintf("%d failed events were detected in the selected audit window.", failedCount))
	}
	if failedCount >= 2 && uniqueIPs == 1 && len(logs) > 0 && strings.TrimSpace(logs[0].IPAddress) != "" {
		signals = append(signals, fmt.Sprintf("Failed activity appears concentrated on a single IP address: %s.", logs[0].IPAddress))
	}
	for _, logEntry := range logs {
		if strings.Contains(strings.ToLower(logEntry.Description), "invalid credentials") {
			signals = append(signals, "Repeated invalid credential events suggest the account should be reviewed for brute-force attempts.")
			break
		}
	}
	if len(signals) == 0 {
		signals = append(signals, "No strong suspicious pattern was found in the selected audit log sample.")
	}
	return signals
}

func buildMockRecommendations(signals []string) []string {
	if len(signals) == 1 && strings.Contains(strings.ToLower(signals[0]), "no strong suspicious pattern") {
		return []string{
			"Keep the audit trail for reference and continue normal monitoring.",
			"Review nearby session activity only if this event range is already under investigation.",
		}
	}
	return []string{
		"Review the affected account together with recent session history and device changes.",
		"Confirm whether the failed or concentrated activity matches an expected test, admin action, or user report.",
	}
}

func mockDescriptionSuffix(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	return fmt.Sprintf("; description: %s", value)
}

func parseMockInt(input string, pattern *regexp.Regexp, group int) int {
	matches := pattern.FindStringSubmatch(input)
	if len(matches) <= group {
		return 0
	}
	return parseMockAtoi(matches[group])
}

func parseMockAtoi(raw string) int {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0
	}
	return value
}

func emptyMockFallback(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
