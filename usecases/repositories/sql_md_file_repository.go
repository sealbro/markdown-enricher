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

type SqlMdFileRepository struct {
	db *db.DB
}

func MakeSqlMdFileRepository(db *db.DB) (repository.MdFileRepository, error) {
	err := db.AutoMigrate(&model.MdFile{})
	return &SqlMdFileRepository{
		db: db,
	}, err
}

func (r *SqlMdFileRepository) Get(ctx context.Context, url string) (*model.MdFile, error) {
	dbModel := &model.MdFile{}
	last := r.db.WithContext(ctx).Last(dbModel, "url = ?", url)
	if errors.Is(last.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return dbModel, last.Error
}

func (r *SqlMdFileRepository) Upsert(ctx context.Context, repoInfo *model.MdFile) error {
	repoInfo.Modified = time.Now()

	columns := []string{"modified", "status"}

	tx := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "url"}},
		DoUpdates: clause.AssignmentColumns(columns),
	}).Create(repoInfo)

	return tx.Error
}
