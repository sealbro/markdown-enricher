package interactors

import (
	"context"
	"errors"
	"markdown-enricher/domain/model"
	"markdown-enricher/infrastructure/cache"
	"markdown-enricher/infrastructure/websocket"
	"markdown-enricher/interfaces/repository"
	"markdown-enricher/pkg/closer"
	"markdown-enricher/pkg/logger"
	"markdown-enricher/usecases/parsers"
	"markdown-enricher/usecases/services"
	"strings"
	"time"
)

type ProcessState struct {
	Url      string `json:"url"`
	Progress int    `json:"progress"`
	Stage    string `json:"stage"`
}

type FilesToProcessQueue chan *model.MdFile

func MakeFilesToProcessQueue() FilesToProcessQueue {
	return make(FilesToProcessQueue, 10)
}

func (q FilesToProcessQueue) Close(context.Context) error {
	close(q)

	return nil
}

type EnricherInteractor struct {
	closed              bool
	parser              *parsers.MarkdownParser
	linkRepository      repository.LinkRepository
	mdFileCache         *cache.TypedCache[*model.MarkdownEnriched]
	repoCache           *cache.TypedCache[*model.GitHubRepoInfo]
	gitHubService       *services.GitHubService
	fileRepository      repository.MdFileRepository
	FilesToProcessQueue FilesToProcessQueue
	wsProvider          *websocket.WebSocketProvider
}

const (
	mdFileCacheDuration  = 24 * time.Hour
	repoCacheDuration    = 2 * 24 * time.Hour
	timeOutMdFileProcess = 10
	timeOutLastRepoInfo  = 24
)

func MakeEnricherInteractor(
	collection *closer.CloserCollection, parser *parsers.MarkdownParser,
	linkStorageRepository repository.LinkRepository, cacheService cache.CacheService,
	gitHubService *services.GitHubService, fileRepository repository.MdFileRepository,
	queue FilesToProcessQueue, provider *websocket.WebSocketProvider) *EnricherInteractor {

	mdFileCache := cache.NewTypedCache[*model.MarkdownEnriched](cacheService, "markdown_file")
	repoCache := cache.NewTypedCache[*model.GitHubRepoInfo](cacheService, "github_repo")

	inretactor := &EnricherInteractor{
		mdFileCache:         mdFileCache,
		repoCache:           repoCache,
		parser:              parser,
		linkRepository:      linkStorageRepository,
		fileRepository:      fileRepository,
		gitHubService:       gitHubService,
		FilesToProcessQueue: queue,
		wsProvider:          provider,
	}

	collection.Add(inretactor)

	return inretactor
}

func (ei *EnricherInteractor) Close(context.Context) error {
	ei.closed = true
	close(ei.FilesToProcessQueue)

	return nil
}

func (ei *EnricherInteractor) Markdown(ctx context.Context, mdFileUrl string) (*model.MarkdownEnriched, error) {
	return ei.mdFileCache.GetOrSet(mdFileUrl, func() (*model.MarkdownEnriched, error) {
		mdFile, err := ei.fileRepository.Get(ctx, mdFileUrl)
		if err != nil || ei.closed {
			return model.EmptyMarkdownEnriched, err
		}

		if mdFile == nil || mdFile.Status != model.Process || time.Now().Sub(mdFile.Modified).Minutes() > timeOutMdFileProcess {
			mdFile = &model.MdFile{
				Created:  time.Now(),
				Modified: time.Now(),
				Url:      mdFileUrl,
				Status:   model.Ready,
			}
			err = ei.fileRepository.Upsert(ctx, mdFile)

			go func() { ei.FilesToProcessQueue <- mdFile }()
		}

		return model.EmptyMarkdownEnriched, err
	}, mdFileCacheDuration)
}

func (ei *EnricherInteractor) ForceMarkdown(ctx context.Context, mdFile *model.MdFile) (*model.MarkdownEnriched, error) {
	state := &ProcessState{Url: mdFile.Url}
	ei.stageNotify(state, "process-start", 0)

	mdFile.Status = model.Process
	err := ei.fileRepository.Upsert(ctx, mdFile)
	if err != nil {
		return nil, err
	}

	mdFileUrl := mdFile.Url

	defer func() {
		logger.Info(ctx, "ForceMarkdown finish: %v", mdFileUrl)
		mdFile.Status = model.Done
		err = ei.fileRepository.Upsert(ctx, mdFile)
		if err != nil {
			logger.Error(ctx, "ForceMarkdown error: %v", err.Error())
		}

		ei.stageNotify(state, "process-done", 100)
	}()

	logger.Info(ctx, "ForceMarkdown start: %v", mdFileUrl)
	linkUrls, err := ei.parser.ExtractLinksFromRemoteFile(ctx, mdFileUrl)
	if err != nil {
		return nil, err
	}
	ei.stageNotify(state, "process-file-parsed", 0)

	links := make([]*model.LinkEnriched, 0)
	countUrls := len(linkUrls)
	// every 10% logging process
	loggingStep := int(float32(countUrls) * 0.1)
	for i, repoUrl := range linkUrls {
		repoInfo, err := ei.repoCache.GetOrSet(repoUrl, ei.getFromStorageOrApi(ctx, repoUrl), repoCacheDuration)
		if err != nil {
			logger.Warn(ctx, "error for [%v]: %v", repoUrl, err.Error())
			continue
		} else {
			links = append(links, &model.LinkEnriched{
				//Url:  repoUrl,
				Info: repoInfo,
			})
		}

		if (i+1)%loggingStep == 0 {
			logger.Info(ctx, "ForceMarkdown processed: [%v/%v] for %v", i, countUrls, mdFileUrl)

			ei.stageNotify(state, "process-github-repository", int(float64(i+1)/float64(countUrls)*100))
		}
	}

	enriched := &model.MarkdownEnriched{Links: links}

	err = ei.mdFileCache.Set(mdFileUrl, enriched, mdFileCacheDuration)
	if err != nil {
		return nil, err
	}

	return enriched, err
}

func (ei *EnricherInteractor) stageNotify(state *ProcessState, stage string, process int) {
	state.Stage = stage
	state.Progress = process
	ei.wsProvider.Send(state)
}

func (ei *EnricherInteractor) getFromStorageOrApi(ctx context.Context, repoUrl string) func() (*model.GitHubRepoInfo, error) {
	return func() (*model.GitHubRepoInfo, error) {
		var err error

		owner, repo, err := splitUrl(repoUrl)
		if err != nil {
			return nil, err
		}

		repoInfo, _ := ei.linkRepository.Get(ctx, owner, repo)
		if repoInfo == nil || time.Now().Sub(repoInfo.Modified).Hours() > timeOutLastRepoInfo {
			repoInfo, err = ei.gitHubService.GetRepoInfo(ctx, owner, repo)
			if err != nil {
				return nil, err
			}

			err = ei.linkRepository.Upsert(ctx, repoInfo)
		}

		logger.Trace(ctx, "Got github info for %v/%v", owner, repo)

		return repoInfo, err
	}
}

func splitUrl(url string) (owner string, repo string, err error) {
	fields := strings.Split(url, "/")

	if len(fields) < 2 {
		return "", "", errors.New("fail after split repo url")
	}

	return fields[0], fields[1], nil
}
