package audit

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

type Filter struct {
	Page        int
	Limit       int
	Action      string
	Resource    string
	Status      string
	ActorUserID *uint
	Search      string
	DateFrom    *time.Time
	DateTo      *time.Time
}

type InvestigationFilter struct {
	Page            int
	Limit           int
	Resource        string
	Status          string
	CreatedByUserID *uint
	AIProvider      string
	AIModel         string
	Search          string
	CreatedFrom     *time.Time
	CreatedTo       *time.Time
}

type Repository interface {
	Create(log *AuditLog) error
	FindAllWithFilter(filter Filter) ([]AuditLog, int64, error)
	FindForExport(filter Filter) ([]AuditLog, error)
	CreateInvestigation(record *AuditInvestigation) error
	FindLatestInvestigationBySnapshot(createdByUserID *uint, snapshotHash string) (*AuditInvestigation, error)
	FindInvestigations(filter InvestigationFilter) ([]AuditInvestigation, int64, error)
	FindInvestigationByID(id uint) (*AuditInvestigation, error)
}

type gormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Create(log *AuditLog) error {
	return r.db.Create(log).Error
}

func (r *gormRepository) CreateInvestigation(record *AuditInvestigation) error {
	err := r.db.Create(record).Error
	if isMissingSnapshotHashColumnError(err) {
		// Backward-compatible fallback for databases that have not applied
		// the snapshot-hash migration yet. This disables dedupe but still
		// lets the investigation be saved.
		return r.db.Omit("SnapshotHash").Create(record).Error
	}
	return err
}

func (r *gormRepository) FindLatestInvestigationBySnapshot(createdByUserID *uint, snapshotHash string) (*AuditInvestigation, error) {
	var item AuditInvestigation

	query := r.db.Where("snapshot_hash = ?", snapshotHash)
	if createdByUserID == nil {
		query = query.Where("created_by_user_id IS NULL")
	} else {
		query = query.Where("created_by_user_id = ?", *createdByUserID)
	}

	if err := query.Order("created_at DESC").First(&item).Error; err != nil {
		if isMissingSnapshotHashColumnError(err) {
			// If the column does not exist yet, treat it as "no reusable record"
			// so investigation creation can proceed without dedupe.
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (r *gormRepository) FindAllWithFilter(filter Filter) ([]AuditLog, int64, error) {
	var (
		logs  []AuditLog
		total int64
	)

	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 10
	}

	query := r.applyFilter(r.db.Model(&AuditLog{}), filter)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.Limit
	err := query.Order("created_at DESC").Limit(filter.Limit).Offset(offset).Find(&logs).Error
	return logs, total, err
}

func (r *gormRepository) FindForExport(filter Filter) ([]AuditLog, error) {
	var logs []AuditLog

	query := r.applyFilter(r.db.Model(&AuditLog{}), filter)
	err := query.Order("created_at DESC").Find(&logs).Error
	return logs, err
}

func (r *gormRepository) FindInvestigations(filter InvestigationFilter) ([]AuditInvestigation, int64, error) {
	var (
		items []AuditInvestigation
		total int64
	)

	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 10
	}

	query := r.applyInvestigationFilter(r.db.Model(&AuditInvestigation{}), filter)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.Limit
	err := query.Order("created_at DESC").Limit(filter.Limit).Offset(offset).Find(&items).Error
	return items, total, err
}

func (r *gormRepository) FindInvestigationByID(id uint) (*AuditInvestigation, error) {
	var item AuditInvestigation
	if err := r.db.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *gormRepository) applyFilter(query *gorm.DB, filter Filter) *gorm.DB {
	if filter.Action != "" {
		query = query.Where("action = ?", filter.Action)
	}
	if filter.Resource != "" {
		query = query.Where("resource = ?", filter.Resource)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.ActorUserID != nil {
		query = query.Where("actor_user_id = ?", *filter.ActorUserID)
	}
	if filter.DateFrom != nil {
		query = query.Where("created_at >= ?", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		query = query.Where("created_at <= ?", *filter.DateTo)
	}
	if search := strings.TrimSpace(filter.Search); search != "" {
		like := "%" + search + "%"
		query = query.Where(
			"(action ILIKE ? OR resource ILIKE ? OR status ILIKE ? OR description ILIKE ? OR ip_address ILIKE ? OR user_agent ILIKE ?)",
			like, like, like, like, like, like,
		)
	}
	return query
}

func (r *gormRepository) applyInvestigationFilter(query *gorm.DB, filter InvestigationFilter) *gorm.DB {
	if filter.Resource != "" {
		query = query.Where("resource = ?", filter.Resource)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.CreatedByUserID != nil {
		query = query.Where("created_by_user_id = ?", *filter.CreatedByUserID)
	}
	if filter.AIProvider != "" {
		query = query.Where("ai_provider = ?", filter.AIProvider)
	}
	if filter.AIModel != "" {
		query = query.Where("ai_model = ?", filter.AIModel)
	}
	if filter.CreatedFrom != nil {
		query = query.Where("created_at >= ?", *filter.CreatedFrom)
	}
	if filter.CreatedTo != nil {
		query = query.Where("created_at <= ?", *filter.CreatedTo)
	}
	if search := strings.TrimSpace(filter.Search); search != "" {
		like := "%" + search + "%"
		query = query.Where(
			"(resource ILIKE ? OR status ILIKE ? OR summary ILIKE ? OR ai_provider ILIKE ? OR ai_model ILIKE ? OR search ILIKE ?)",
			like, like, like, like, like, like,
		)
	}
	return query
}

func isMissingSnapshotHashColumnError(err error) bool {
	if err == nil {
		return false
	}

	message := strings.ToLower(err.Error())
	return strings.Contains(message, "snapshot_hash") &&
		(strings.Contains(message, "column") ||
			strings.Contains(message, "unknown column") ||
			strings.Contains(message, "does not exist"))
}
