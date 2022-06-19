package repository

import (
	"context"
	"markdown-enricher/domain/model"
)

type MdFileRepository interface {
	Get(ctx context.Context, url string) (*model.MdFile, error)
	Upsert(ctx context.Context, repoInfo *model.MdFile) error
}
