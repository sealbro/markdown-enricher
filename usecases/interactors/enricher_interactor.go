package interactors

import (
	"context"
	"markdown-enricher/domain/model"
	"markdown-enricher/infrastructure/cache"
	"markdown-enricher/interfaces/repository"
	"markdown-enricher/usecases/parsers"
	"markdown-enricher/usecases/services"
	"strings"
	"time"
)

type EnricherInteractor struct {
	parser                *parsers.MarkdownParser
	linkStorageRepository repository.LinkStorageRepository
	mdFileCache           *cache.TypedCache[*model.MarkdownEnriched]
	repoCache             *cache.TypedCache[*model.GitHubRepoInfo]
	gitHubService         *services.GitHubService
}

func MakeEnricherInteractor(parser *parsers.MarkdownParser, linkStorageRepository repository.LinkStorageRepository, cacheService cache.CacheService, gitHubService *services.GitHubService) *EnricherInteractor {
	mdFileCache := cache.NewTypedCache[*model.MarkdownEnriched](cacheService, "markdown_file=>")
	repoCache := cache.NewTypedCache[*model.GitHubRepoInfo](cacheService, "github_repo=>")

	return &EnricherInteractor{
		mdFileCache:           mdFileCache,
		repoCache:             repoCache,
		parser:                parser,
		linkStorageRepository: linkStorageRepository,
		gitHubService:         gitHubService,
	}
}

func (ei *EnricherInteractor) Markdown(ctx context.Context, mdFileUrl string) (*model.MarkdownEnriched, error) {
	enriched, err := ei.mdFileCache.GetOrSet(mdFileUrl, func() (*model.MarkdownEnriched, error) {
		linkUrls, err := ei.parser.ExtractLinksFromRemoteFile(ctx, mdFileUrl)
		if err != err {
			return nil, err
		}

		links := make([]*model.LinkEnriched, len(linkUrls))
		for _, repoUrl := range linkUrls {
			repoInfo, err := ei.repoCache.GetOrSet(repoUrl, ei.getFromStorageOrApi(ctx, repoUrl), 24*time.Hour)
			if err != nil {
				continue
			} else {
				links = append(links, &model.LinkEnriched{
					Url:  repoUrl,
					Info: repoInfo,
				})
			}
		}

		return &model.MarkdownEnriched{Links: nil}, nil
	}, 3*24*time.Hour)

	return enriched, err
}

func (ei *EnricherInteractor) getFromStorageOrApi(ctx context.Context, repoUrl string) func() (*model.GitHubRepoInfo, error) {
	return func() (*model.GitHubRepoInfo, error) {
		var err error

		owner, repo := splitUrl(repoUrl)
		repoInfo, _ := ei.linkStorageRepository.Get(ctx, owner, repo)
		if repoInfo == nil || time.Now().Sub(repoInfo.Modified).Hours() > 24 {
			repoInfo, err = ei.gitHubService.GetRepoInfo(ctx, owner, repo)
			if err != nil {
				return nil, err
			}

			err = ei.linkStorageRepository.Upsert(ctx, repoInfo)
		}

		return repoInfo, err
	}
}

func splitUrl(url string) (owner string, repo string) {
	fields := strings.Split(url, "/")

	return fields[len(fields)-2], fields[len(fields)-1]
}
