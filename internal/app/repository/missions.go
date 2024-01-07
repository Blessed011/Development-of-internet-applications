package repository

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"lab1/internal/app/ds"
)

func (r *Repository) GetAllMissions(customerId *string, formationDateStart, formationCompletionDate *time.Time, status string) ([]ds.Mission, error) {
	var missions []ds.Mission

	query := r.db.Preload("Customer").Preload("Moderator").
		Where("LOWER(status) LIKE ?", "%"+strings.ToLower(status)+"%").
		Where("status != ? AND status != ?", ds.StatusDeleted, ds.StatusDraft)

	if customerId != nil {
		query = query.Where("customer_id = ?", *customerId)
	}
	if formationDateStart != nil && formationCompletionDate != nil {
		query = query.Where("formation_date BETWEEN ? AND ?", *formationDateStart, *formationCompletionDate)
	} else if formationDateStart != nil {
		query = query.Where("formation_date >= ?", *formationDateStart)
	} else if formationCompletionDate != nil {
		query = query.Where("formation_date <= ?", *formationCompletionDate)
	}

	if err := query.Find(&missions).Error; err != nil {
		return nil, err
	}
	return missions, nil
}

func (r *Repository) GetDraftMission(customerId string) (*ds.Mission, error) {
	mission := &ds.Mission{}
	err := r.db.First(mission, ds.Mission{Status: ds.StatusDraft, CustomerId: customerId}).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mission, nil
}

func (r *Repository) CreateDraftMission(customerId string) (*ds.Mission, error) {
	mission := &ds.Mission{CreationDate: time.Now(), CustomerId: customerId, Status: ds.StatusDraft}
	err := r.db.Create(mission).Error
	if err != nil {
		return nil, err
	}
	return mission, nil
}

func (r *Repository) GetMissionById(missionId string, userId *string) (*ds.Mission, error) {
	mission := &ds.Mission{}
	query := r.db.Preload("Moderator").Preload("Customer").
		Where("status != ?", ds.StatusDeleted)
	if userId != nil {
		query = query.Where("customer_id = ?", userId)
	}
	err := query.First(mission, ds.Mission{UUID: missionId}).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mission, nil
}

func (r *Repository) GetFlight(missionId string) ([]ds.Module, error) {
	var modules []ds.Module

	err := r.db.Table("flights").
		Select("modules.*").
		Joins("JOIN modules ON flights.module_id = modules.uuid").
		Where(ds.Flight{MissionId: missionId}).
		Scan(&modules).Error

	if err != nil {
		return nil, err
	}
	return modules, nil
}

func (r *Repository) SaveMission(mission *ds.Mission) error {
	err := r.db.Save(mission).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) DeleteFromMission(missionId, moduleId string) error {
	err := r.db.Delete(&ds.Flight{MissionId: missionId, ModuleId: moduleId}).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) CountModules(missionId string) (int64, error) {
	var count int64
	err := r.db.Model(&ds.Flight{}).
		Where("mission_id = ?", missionId).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}
