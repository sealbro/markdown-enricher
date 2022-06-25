package services

import (
	"context"
	"fmt"
	"github.com/google/go-github/v45/github"
	"markdown-enricher/domain/model"
	"os"
	"time"
)
import "golang.org/x/oauth2"

type GitHubService struct {
	client *github.Client
}

func MakeGitHubService() (*GitHubService, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	background := context.Background()
	tc := oauth2.NewClient(background, ts)

	client := github.NewClient(tc)

	err := checkToken(background, client)

	return &GitHubService{
		client: client,
	}, err
}

func checkToken(background context.Context, client *github.Client) error {
	ctx, cancelFunc := context.WithTimeout(background, time.Second)
	defer cancelFunc()
	_, _, err := client.Activity.IsStarred(ctx, "avelino", "awesome-go")
	if err != nil {
		return fmt.Errorf("wrong GITHUB_TOKEN: %w", err)
	}

	return nil
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
