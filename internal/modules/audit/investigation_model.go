package audit

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type AuditInvestigation struct {
	gorm.Model
	CreatedByUserID       *uint      `json:"created_by_user_id"`
	SnapshotHash          string     `json:"-"`
	Action                string     `json:"action"`
	Resource              string     `json:"resource"`
	Status                string     `json:"status"`
	ActorUserID           *uint      `json:"actor_user_id"`
	Search                string     `json:"search"`
	DateFrom              *time.Time `json:"date_from"`
	DateTo                *time.Time `json:"date_to"`
	LimitValue            int        `json:"limit"`
	LogCount              int        `json:"log_count"`
	AIProvider            string     `json:"ai_provider"`
	AIModel               string     `json:"ai_model"`
	Summary               string     `json:"summary"`
	TimelineJSON          string     `json:"-"`
	SuspiciousSignalsJSON string     `json:"-"`
	RecommendationsJSON   string     `json:"-"`
}

func (AuditInvestigation) TableName() string {
	return "audit_investigations"
}

type InvestigationHistory struct {
	ID                uint       `json:"id"`
	CreatedAt         time.Time  `json:"created_at"`
	CreatedByUserID   *uint      `json:"created_by_user_id"`
	Action            string     `json:"action"`
	Resource          string     `json:"resource"`
	Status            string     `json:"status"`
	ActorUserID       *uint      `json:"actor_user_id"`
	Search            string     `json:"search"`
	DateFrom          *time.Time `json:"date_from"`
	DateTo            *time.Time `json:"date_to"`
	Limit             int        `json:"limit"`
	LogCount          int        `json:"log_count"`
	AIProvider        string     `json:"ai_provider"`
	AIModel           string     `json:"ai_model"`
	Summary           string     `json:"summary"`
	Timeline          []string   `json:"timeline"`
	SuspiciousSignals []string   `json:"suspicious_signals"`
	Recommendations   []string   `json:"recommendations"`
}

func (m AuditInvestigation) ToHistory() InvestigationHistory {
	return InvestigationHistory{
		ID:                m.ID,
		CreatedAt:         m.CreatedAt.UTC(),
		CreatedByUserID:   m.CreatedByUserID,
		Action:            m.Action,
		Resource:          m.Resource,
		Status:            m.Status,
		ActorUserID:       m.ActorUserID,
		Search:            m.Search,
		DateFrom:          m.DateFrom,
		DateTo:            m.DateTo,
		Limit:             m.LimitValue,
		LogCount:          m.LogCount,
		AIProvider:        m.AIProvider,
		AIModel:           m.AIModel,
		Summary:           m.Summary,
		Timeline:          decodeStringSlice(m.TimelineJSON),
		SuspiciousSignals: decodeStringSlice(m.SuspiciousSignalsJSON),
		Recommendations:   decodeStringSlice(m.RecommendationsJSON),
	}
}

func encodeStringSlice(items []string) string {
	if len(items) == 0 {
		return "[]"
	}
	payload, err := json.Marshal(items)
	if err != nil {
		return "[]"
	}
	return string(payload)
}

func decodeStringSlice(raw string) []string {
	if raw == "" {
		return []string{}
	}

	var items []string
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return []string{}
	}
	return items
}
