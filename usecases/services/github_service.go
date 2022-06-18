package services

import (
	"context"
	"github.com/google/go-github/v45/github"
	"markdown-enricher/domain/model"
	"os"
	"time"
)
import "golang.org/x/oauth2"

type GitHubService struct {
	client *github.Client
}

func MakeGitHubService() *GitHubService {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	tc := oauth2.NewClient(context.Background(), ts)

	client := github.NewClient(tc)

	return &GitHubService{
		client: client,
	}
}

func (s *GitHubService) GetRepoInfo(ctx context.Context, owner, repo string) (*model.GitHubRepoInfo, error) {
	repository, _, err := s.client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	var starsCount int
	if repository.StargazersCount != nil {
		starsCount = *repository.StargazersCount
	}

	var forkCount int
	if repository.ForksCount != nil {
		forkCount = *repository.ForksCount
	}

	return &model.GitHubRepoInfo{
		Created:    time.Now(),
		Modified:   time.Now(),
		Owner:      owner,
		Repo:       repo,
		Stars:      starsCount,
		Forks:      forkCount,
		LastCommit: (*repository.PushedAt).Time,
	}, nil
}
