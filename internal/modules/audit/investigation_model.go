package audit

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type AuditInvestigation struct {
	gorm.Model
	CreatedByUserID       *uint      `json:"created_by_user_id" gorm:"column:created_by_user_id"`
	SnapshotHash          string     `json:"-" gorm:"column:snapshot_hash"`
	Action                string     `json:"action" gorm:"column:action"`
	Resource              string     `json:"resource" gorm:"column:resource"`
	Status                string     `json:"status" gorm:"column:status"`
	ActorUserID           *uint      `json:"actor_user_id" gorm:"column:actor_user_id"`
	Search                string     `json:"search" gorm:"column:search"`
	DateFrom              *time.Time `json:"date_from" gorm:"column:date_from"`
	DateTo                *time.Time `json:"date_to" gorm:"column:date_to"`
	LimitValue            int        `json:"limit" gorm:"column:limit_value"`
	LogCount              int        `json:"log_count" gorm:"column:log_count"`
	AIProvider            string     `json:"ai_provider" gorm:"column:ai_provider"`
	AIModel               string     `json:"ai_model" gorm:"column:ai_model"`
	Summary               string     `json:"summary" gorm:"column:summary"`
	TimelineJSON          string     `json:"-" gorm:"column:timeline_json"`
	SuspiciousSignalsJSON string     `json:"-" gorm:"column:suspicious_signals_json"`
	RecommendationsJSON   string     `json:"-" gorm:"column:recommendations_json"`
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
