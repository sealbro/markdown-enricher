package repositories

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"markdown-enricher/domain/model"
	"markdown-enricher/infrastructure/db"
	"markdown-enricher/interfaces/repository"
)

type SqlLinkStorageRepository struct {
	db *db.DB
}

func MakeSqlStorageRepository(db *db.DB) repository.LinkStorageRepository {
	db.AutoMigrate(&model.GitHubRepoInfo{})
	return &SqlLinkStorageRepository{
		db: db,
	}
}

func (r *SqlLinkStorageRepository) Get(ctx context.Context, owner, repo string) (*model.GitHubRepoInfo, error) {
	state := &model.GitHubRepoInfo{}
	last := r.db.WithContext(ctx).Last(state, "owner = ? and repo = ?", owner, repo)
	if errors.Is(last.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return state, last.Error
}

func (r *SqlLinkStorageRepository) Upsert(ctx context.Context, repoInfo *model.GitHubRepoInfo) error {
	columns := []string{"modified", "stars", "forks", "last_commit"}

	tx := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "owner"}, {Name: "repo"}},
		DoUpdates: clause.AssignmentColumns(columns),
	}).Create(repoInfo)

	return tx.Error
}
