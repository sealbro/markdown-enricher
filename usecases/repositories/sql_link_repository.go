package repositories

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"markdown-enricher/domain/model"
	"markdown-enricher/infrastructure/db"
	"markdown-enricher/interfaces/repository"
	"time"
)

type SqlLinkRepository struct {
	db *db.DB
}

func MakeSqlLinkRepository(db *db.DB) (repository.LinkRepository, error) {
	err := db.AutoMigrate(&model.GitHubRepoInfo{})
	return &SqlLinkRepository{
		db: db,
	}, err
}

func (r *SqlLinkRepository) Get(ctx context.Context, owner, repo string) (*model.GitHubRepoInfo, error) {
	dbModel := &model.GitHubRepoInfo{}
	last := r.db.WithContext(ctx).Last(dbModel, "owner = ? and repo = ?", owner, repo)
	if errors.Is(last.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return dbModel, last.Error
}

func (r *SqlLinkRepository) Upsert(ctx context.Context, repoInfo *model.GitHubRepoInfo) error {
	repoInfo.Modified = time.Now()

	columns := []string{"modified", "stars", "forks", "last_commit"}

	tx := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "owner"}, {Name: "repo"}},
		DoUpdates: clause.AssignmentColumns(columns),
	}).Create(repoInfo)

	return tx.Error
}
