package repository

import (
	"context"
	"markdown-enricher/domain/model"
)

type LinkRepository interface {
	Get(ctx context.Context, owner, repo string) (*model.GitHubRepoInfo, error)
	Upsert(ctx context.Context, repoInfo *model.GitHubRepoInfo) error
}
