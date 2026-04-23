package audit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"go-api-starterkit/internal/ai"

	"gorm.io/gorm"
)

type InvestigatorService struct {
	AI       *ai.Service
	Repo     Repository
	AuditLog *Service
}

const (
	investigationPromptVersion  = "v3"
	investigationRawLogLimit    = 25
	investigationChunkSize      = 20
	investigationChunkSampleCap = 3
)

func NewInvestigatorService(repo Repository, aiService *ai.Service, auditLogService *Service) *InvestigatorService {
	return &InvestigatorService{
		AI:       aiService,
		Repo:     repo,
		AuditLog: auditLogService,
	}
}

func (s *InvestigatorService) Investigate(ctx context.Context, filter Filter) (*InvestigationResult, []AuditLog, error) {
	if s == nil || s.AI == nil || !s.AI.Enabled() {
		return nil, nil, errors.New("ai investigator is not enabled")
	}

	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	filter.Page = 1

	logs, _, err := s.Repo.FindAllWithFilter(filter)
	if err != nil {
		return nil, nil, err
	}
	if len(logs) == 0 {
		return nil, nil, errors.New("no audit logs found for investigation")
	}

	input := ai.BuildJSONPrompt(
		"Summarize these audit logs into the required JSON schema. Mention actor_user_id only when it is present. Mention ip_address separately when relevant. Keep timeline and suspicious_signals concise and human-readable.",
		buildInvestigationContext(logs),
	)
	result, err := s.AI.Generate(ctx, input)
	if err != nil {
		return nil, nil, err
	}

	parsed, err := parseInvestigationResult(strings.TrimSpace(result.Text))
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ai.ErrInvalidStructuredOutput, err)
	}
	parsed = mergeInvestigationHeuristics(parsed, logs)

	return parsed, logs, nil
}

func (s *InvestigatorService) SaveInvestigation(createdByUserID *uint, filter Filter, result *InvestigationResult, logs []AuditLog, ipAddress string, userAgent string) (*InvestigationHistory, bool, error) {
	if s == nil || s.Repo == nil {
		return nil, false, errors.New("audit repository is not configured")
	}

	snapshotHash, err := buildInvestigationSnapshotHash(createdByUserID, filter, logs, s.AI)
	if err != nil {
		return nil, false, err
	}

	existing, err := s.Repo.FindLatestInvestigationBySnapshot(createdByUserID, snapshotHash)
	if err == nil && existing != nil {
		history := existing.ToHistory()
		s.recordInvestigationAudit(createdByUserID, history.ID, filter, history.LogCount, "reused", ipAddress, userAgent)
		return &history, true, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, err
	}

	record := &AuditInvestigation{
		CreatedByUserID:       createdByUserID,
		SnapshotHash:          snapshotHash,
		Action:                filter.Action,
		Resource:              filter.Resource,
		Status:                filter.Status,
		ActorUserID:           filter.ActorUserID,
		Search:                filter.Search,
		DateFrom:              filter.DateFrom,
		DateTo:                filter.DateTo,
		LimitValue:            filter.Limit,
		LogCount:              len(logs),
		AIProvider:            s.AI.ProviderName(),
		AIModel:               s.AI.ModelName(),
		Summary:               result.Summary,
		TimelineJSON:          encodeStringSlice(result.Timeline),
		SuspiciousSignalsJSON: encodeStringSlice(result.SuspiciousSignals),
		RecommendationsJSON:   encodeStringSlice(result.Recommendations),
	}

	if err := s.Repo.CreateInvestigation(record); err != nil {
		return nil, false, err
	}

	history := record.ToHistory()
	s.recordInvestigationAudit(createdByUserID, history.ID, filter, len(logs), "created", ipAddress, userAgent)
	return &history, false, nil
}

func (s *InvestigatorService) ListInvestigations(filter InvestigationFilter) ([]InvestigationHistory, int64, error) {
	if s == nil || s.Repo == nil {
		return nil, 0, errors.New("audit repository is not configured")
	}

	records, total, err := s.Repo.FindInvestigations(filter)
	if err != nil {
		return nil, 0, err
	}

	items := make([]InvestigationHistory, 0, len(records))
	for _, record := range records {
		items = append(items, record.ToHistory())
	}
	return items, total, nil
}

func (s *InvestigatorService) GetInvestigationByID(id uint) (*InvestigationHistory, error) {
	if s == nil || s.Repo == nil {
		return nil, errors.New("audit repository is not configured")
	}

	record, err := s.Repo.FindInvestigationByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}

	history := record.ToHistory()
	return &history, nil
}

func buildInvestigationSnapshotHash(createdByUserID *uint, filter Filter, logs []AuditLog, aiService *ai.Service) (string, error) {
	logIDs := make([]uint, 0, len(logs))
	for _, logEntry := range logs {
		logIDs = append(logIDs, logEntry.ID)
	}

	payload := struct {
		PromptVersion   string `json:"prompt_version"`
		CreatedByUserID *uint  `json:"created_by_user_id,omitempty"`
		Action          string `json:"action"`
		Resource        string `json:"resource"`
		Status          string `json:"status"`
		ActorUserID     *uint  `json:"actor_user_id,omitempty"`
		Search          string `json:"search"`
		DateFrom        string `json:"date_from,omitempty"`
		DateTo          string `json:"date_to,omitempty"`
		Limit           int    `json:"limit"`
		LogIDs          []uint `json:"log_ids"`
		AIProvider      string `json:"ai_provider"`
		AIModel         string `json:"ai_model"`
	}{
		PromptVersion:   investigationPromptVersion,
		CreatedByUserID: createdByUserID,
		Action:          filter.Action,
		Resource:        filter.Resource,
		Status:          filter.Status,
		ActorUserID:     filter.ActorUserID,
		Search:          strings.TrimSpace(filter.Search),
		DateFrom:        formatSnapshotTime(filter.DateFrom),
		DateTo:          formatSnapshotTime(filter.DateTo),
		Limit:           filter.Limit,
		LogIDs:          logIDs,
		AIProvider:      providerName(aiService),
		AIModel:         modelName(aiService),
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:]), nil
}

func formatSnapshotTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func providerName(aiService *ai.Service) string {
	if aiService == nil {
		return ""
	}
	return aiService.ProviderName()
}

func modelName(aiService *ai.Service) string {
	if aiService == nil {
		return ""
	}
	return aiService.ModelName()
}

func (s *InvestigatorService) recordInvestigationAudit(createdByUserID *uint, investigationID uint, filter Filter, logCount int, action string, ipAddress string, userAgent string) {
	if s == nil || s.AuditLog == nil {
		return
	}

	resourceID := investigationID
	description := fmt.Sprintf(
		"audit investigation %s for resource=%s action=%s status=%s log_count=%d",
		action,
		emptyStringFallback(filter.Resource),
		emptyStringFallback(filter.Action),
		emptyStringFallback(filter.Status),
		logCount,
	)

	s.AuditLog.SafeRecord(RecordInput{
		ActorUserID: createdByUserID,
		Action:      action,
		Resource:    "audit_investigation",
		ResourceID:  &resourceID,
		Status:      "success",
		Description: description,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
	})
}

func buildInvestigationContext(logs []AuditLog) string {
	lines := make([]string, 0, len(logs)+4)
	lines = append(lines, fmt.Sprintf("Audit log count: %d", len(logs)))
	lines = append(lines, buildInvestigationOverview(logs))
	lines = append(lines, "Use the event summaries below to build a concise timeline and suspicious signals.")

	if len(logs) > investigationRawLogLimit {
		lines = append(lines, fmt.Sprintf(
			"Context mode: chunked_summary. The %d logs were compressed into %d chunks of at most %d logs each before sending to AI.",
			len(logs),
			chunkCount(len(logs), investigationChunkSize),
			investigationChunkSize,
		))
		lines = append(lines, buildInvestigationChunkSummaries(logs)...)
		return strings.Join(lines, "\n")
	}

	lines = append(lines, buildDetailedInvestigationLogs(logs)...)
	return strings.Join(lines, "\n")
}

func buildDetailedInvestigationLogs(logs []AuditLog) []string {
	lines := make([]string, 0, len(logs)*10)
	for _, logEntry := range logs {
		lines = append(lines, fmt.Sprintf("- log %d", logEntry.ID))
		lines = append(lines, fmt.Sprintf("  time: %s", logEntry.CreatedAt.UTC().Format(time.RFC3339)))
		lines = append(lines, fmt.Sprintf("  action: %s", logEntry.Action))
		lines = append(lines, fmt.Sprintf("  resource: %s", logEntry.Resource))
		lines = append(lines, fmt.Sprintf("  status: %s", logEntry.Status))
		lines = append(lines, fmt.Sprintf("  actor_user_id: %s", pointerStringFallback(logEntry.ActorUserID)))
		lines = append(lines, fmt.Sprintf("  resource_id: %s", pointerStringFallback(logEntry.ResourceID)))
		lines = append(lines, fmt.Sprintf("  ip_address: %s", emptyStringFallback(logEntry.IPAddress)))
		lines = append(lines, fmt.Sprintf("  user_agent: %s", emptyStringFallback(logEntry.UserAgent)))
		lines = append(lines, fmt.Sprintf("  description: %q", strings.TrimSpace(logEntry.Description)))
	}
	return lines
}

func buildInvestigationChunkSummaries(logs []AuditLog) []string {
	lines := make([]string, 0, chunkCount(len(logs), investigationChunkSize)*6)
	for chunkIndex, start := 0, 0; start < len(logs); chunkIndex, start = chunkIndex+1, start+investigationChunkSize {
		end := start + investigationChunkSize
		if end > len(logs) {
			end = len(logs)
		}
		chunk := logs[start:end]
		lines = append(lines, summarizeInvestigationChunk(chunkIndex+1, chunk)...)
	}
	return lines
}

func summarizeInvestigationChunk(chunkNumber int, logs []AuditLog) []string {
	if len(logs) == 0 {
		return nil
	}

	failedCount := 0
	uniqueIPs := make(map[string]struct{})
	uniqueActors := make(map[string]struct{})
	actionCounts := make(map[string]int)
	firstSeen := logs[0].CreatedAt.UTC()
	lastSeen := logs[0].CreatedAt.UTC()

	lines := []string{
		fmt.Sprintf("- chunk %d", chunkNumber),
	}

	for _, logEntry := range logs {
		if strings.EqualFold(logEntry.Status, "failed") {
			failedCount++
		}
		if ip := strings.TrimSpace(logEntry.IPAddress); ip != "" {
			uniqueIPs[ip] = struct{}{}
		}
		if actor := pointerStringFallback(logEntry.ActorUserID); actor != "n/a" {
			uniqueActors[actor] = struct{}{}
		}
		actionKey := fmt.Sprintf("%s/%s/%s", emptyStringFallback(logEntry.Status), emptyStringFallback(logEntry.Action), emptyStringFallback(logEntry.Resource))
		actionCounts[actionKey]++
		if ts := logEntry.CreatedAt.UTC(); ts.Before(firstSeen) {
			firstSeen = ts
		}
		if ts := logEntry.CreatedAt.UTC(); ts.After(lastSeen) {
			lastSeen = ts
		}
	}

	lines = append(lines,
		fmt.Sprintf("  range: %s to %s", firstSeen.Format(time.RFC3339), lastSeen.Format(time.RFC3339)),
		fmt.Sprintf("  totals: logs=%d failed=%d unique_ip_addresses=%d unique_actor_user_ids=%d", len(logs), failedCount, len(uniqueIPs), len(uniqueActors)),
		fmt.Sprintf("  top_patterns: %s", strings.Join(topChunkPatterns(actionCounts, 3), "; ")),
	)

	sampleLimit := investigationChunkSampleCap
	if len(logs) < sampleLimit {
		sampleLimit = len(logs)
	}
	for _, logEntry := range logs[:sampleLimit] {
		lines = append(lines, fmt.Sprintf(
			"  sample_event: %s - %s %s on %s from ip %s (actor_user_id: %s)",
			logEntry.CreatedAt.UTC().Format(time.RFC3339),
			emptyStringFallback(logEntry.Status),
			emptyStringFallback(logEntry.Action),
			emptyStringFallback(logEntry.Resource),
			emptyStringFallback(logEntry.IPAddress),
			pointerStringFallback(logEntry.ActorUserID),
		))
	}

	return lines
}

func topChunkPatterns(counts map[string]int, limit int) []string {
	type item struct {
		label string
		count int
	}

	items := make([]item, 0, len(counts))
	for label, count := range counts {
		items = append(items, item{label: label, count: count})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].count == items[j].count {
			return items[i].label < items[j].label
		}
		return items[i].count > items[j].count
	})

	if len(items) > limit {
		items = items[:limit]
	}

	result := make([]string, 0, len(items))
	for _, item := range items {
		result = append(result, fmt.Sprintf("%s x%d", item.label, item.count))
	}
	return result
}

func chunkCount(total int, size int) int {
	if total <= 0 || size <= 0 {
		return 0
	}
	count := total / size
	if total%size != 0 {
		count++
	}
	return count
}

func buildInvestigationOverview(logs []AuditLog) string {
	failedCount := 0
	uniqueIPs := make(map[string]struct{})
	uniqueActors := make(map[string]struct{})
	firstSeen := logs[0].CreatedAt.UTC()
	lastSeen := logs[0].CreatedAt.UTC()

	for _, logEntry := range logs {
		if strings.EqualFold(logEntry.Status, "failed") {
			failedCount++
		}
		if ip := strings.TrimSpace(logEntry.IPAddress); ip != "" {
			uniqueIPs[ip] = struct{}{}
		}
		if actor := pointerStringFallback(logEntry.ActorUserID); actor != "n/a" {
			uniqueActors[actor] = struct{}{}
		}
		if ts := logEntry.CreatedAt.UTC(); ts.Before(firstSeen) {
			firstSeen = ts
		}
		if ts := logEntry.CreatedAt.UTC(); ts.After(lastSeen) {
			lastSeen = ts
		}
	}

	return fmt.Sprintf(
		"Overview: failed_logs=%d unique_ip_addresses=%d unique_actor_user_ids=%d first_seen=%s last_seen=%s",
		failedCount,
		len(uniqueIPs),
		len(uniqueActors),
		firstSeen.Format(time.RFC3339),
		lastSeen.Format(time.RFC3339),
	)
}

func parseInvestigationResult(raw string) (*InvestigationResult, error) {
	raw = extractJSONObject(raw)

	var payload map[string]any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, err
	}

	return &InvestigationResult{
		Summary:           stringifyValue(payload["summary"]),
		Timeline:          stringifyList(payload["timeline"]),
		SuspiciousSignals: stringifyList(payload["suspicious_signals"]),
		Recommendations:   stringifyList(payload["recommendations"]),
	}, nil
}

func extractJSONObject(raw string) string {
	raw = sanitizeAIResponse(raw)
	for start := 0; start < len(raw); start++ {
		if raw[start] != '{' {
			continue
		}

		depth := 0
		inString := false
		escaped := false
		for i := start; i < len(raw); i++ {
			ch := raw[i]

			if inString {
				if escaped {
					escaped = false
					continue
				}
				if ch == '\\' {
					escaped = true
					continue
				}
				if ch == '"' {
					inString = false
				}
				continue
			}

			switch ch {
			case '"':
				inString = true
			case '{':
				depth++
			case '}':
				depth--
				if depth == 0 {
					candidate := strings.TrimSpace(raw[start : i+1])
					if json.Valid([]byte(candidate)) {
						return candidate
					}
				}
			}
		}
	}

	return raw
}

func sanitizeAIResponse(raw string) string {
	replacer := strings.NewReplacer(
		"```json", "",
		"```JSON", "",
		"```", "",
	)
	return strings.TrimSpace(replacer.Replace(raw))
}

func stringifyList(value any) []string {
	if value == nil {
		return nil
	}

	items, ok := value.([]any)
	if !ok {
		return []string{stringifyValue(value)}
	}

	result := make([]string, 0, len(items))
	for _, item := range items {
		text := stringifyValue(item)
		if text != "" {
			result = append(result, text)
		}
	}
	return result
}

func stringifyValue(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case map[string]any:
		return humanizeMap(v)
	default:
		bytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(bytes)
	}
}

func humanizeMap(value map[string]any) string {
	preferredOrder := []string{
		"time",
		"timestamp",
		"action",
		"event_type",
		"resource",
		"status",
		"actor_user_id",
		"actor",
		"ip_address",
		"details",
		"pattern",
		"total_occurrences",
	}

	parts := make([]string, 0, len(value))
	seen := make(map[string]struct{}, len(preferredOrder))
	for _, key := range preferredOrder {
		if raw, ok := value[key]; ok {
			if text := formatKeyValue(key, raw); text != "" {
				parts = append(parts, text)
			}
			seen[key] = struct{}{}
		}
	}
	for key, raw := range value {
		if _, ok := seen[key]; ok {
			continue
		}
		if text := formatKeyValue(key, raw); text != "" {
			parts = append(parts, text)
		}
	}

	return strings.Join(parts, "; ")
}

func formatKeyValue(key string, value any) string {
	text := strings.TrimSpace(stringifyScalar(value))
	if text == "" {
		return ""
	}
	return fmt.Sprintf("%s: %s", key, text)
}

func stringifyScalar(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case float64:
		if v == float64(int64(v)) {
			return fmt.Sprintf("%.0f", v)
		}
		return fmt.Sprintf("%v", v)
	case bool:
		return fmt.Sprintf("%t", v)
	default:
		bytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(bytes)
	}
}

func emptyStringFallback(value string) string {
	if strings.TrimSpace(value) == "" {
		return "n/a"
	}
	return value
}

func pointerStringFallback(value *uint) string {
	text := strings.TrimSpace(uintPointerString(value))
	if text == "" {
		return "n/a"
	}
	return text
}

func mergeInvestigationHeuristics(result *InvestigationResult, logs []AuditLog) *InvestigationResult {
	if result == nil {
		result = &InvestigationResult{}
	}
	if len(logs) == 0 {
		return result
	}

	if needsTimelineFallback(result.Timeline) {
		result.Timeline = summarizeLogsForTimeline(logs)
	}

	heuristicSignals := deriveHeuristicSignals(logs)
	result.SuspiciousSignals = mergeUniqueLines(result.SuspiciousSignals, heuristicSignals)

	if len(result.Recommendations) == 0 {
		result.Recommendations = deriveRecommendations(logs)
	}

	return result
}

func needsTimelineFallback(lines []string) bool {
	if len(lines) == 0 {
		return true
	}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if _, err := time.Parse(time.RFC3339, trimmed); err != nil {
			return false
		}
	}
	return true
}

func summarizeLogsForTimeline(logs []AuditLog) []string {
	limit := min(len(logs), 5)
	lines := make([]string, 0, limit)
	for _, logEntry := range logs[:limit] {
		lines = append(lines, fmt.Sprintf(
			"%s - %s %s on %s from ip %s (actor_user_id: %s)",
			logEntry.CreatedAt.UTC().Format(time.RFC3339),
			logEntry.Status,
			logEntry.Action,
			logEntry.Resource,
			emptyStringFallback(logEntry.IPAddress),
			pointerStringFallback(logEntry.ActorUserID),
		))
	}
	return lines
}

func deriveHeuristicSignals(logs []AuditLog) []string {
	failedByIP := make(map[string]int)
	failedByAction := make(map[string]int)
	var signals []string

	for _, logEntry := range logs {
		if !strings.EqualFold(logEntry.Status, "failed") {
			continue
		}
		ip := emptyStringFallback(logEntry.IPAddress)
		failedByIP[ip]++
		key := strings.TrimSpace(logEntry.Action + "|" + logEntry.Resource)
		failedByAction[key]++
	}

	for ip, count := range failedByIP {
		if count >= 2 && ip != "n/a" {
			signals = append(signals, fmt.Sprintf("%d failed events originated from ip %s within the selected range.", count, ip))
		}
	}
	for key, count := range failedByAction {
		if count >= 2 {
			parts := strings.SplitN(key, "|", 2)
			signals = append(signals, fmt.Sprintf("%d failed %s events were recorded on resource %s.", count, parts[0], parts[1]))
		}
	}

	sort.Strings(signals)
	return signals
}

func deriveRecommendations(logs []AuditLog) []string {
	signals := deriveHeuristicSignals(logs)
	if len(signals) == 0 {
		return []string{"Review the raw audit logs together with related user sessions before closing the investigation."}
	}
	return []string{
		"Review the affected account and confirm whether the failed activity was expected.",
		"Cross-check the suspicious events with session history and recent permission changes.",
	}
}

func mergeUniqueLines(existing []string, extras []string) []string {
	seen := make(map[string]struct{}, len(existing)+len(extras))
	merged := make([]string, 0, len(existing)+len(extras))

	appendLine := func(line string) {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			return
		}
		if strings.EqualFold(trimmed, "none detected.") && len(extras) > 0 {
			return
		}
		if _, ok := seen[trimmed]; ok {
			return
		}
		seen[trimmed] = struct{}{}
		merged = append(merged, trimmed)
	}

	for _, line := range existing {
		appendLine(line)
	}
	for _, line := range extras {
		appendLine(line)
	}

	return merged
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
