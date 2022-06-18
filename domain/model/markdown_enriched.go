package model

type LinkEnriched struct {
	Url  string          `json:"repo_url"`
	Info *GitHubRepoInfo `json:"info"`
}

type MarkdownEnriched struct {
	Links []*LinkEnriched `json:"links"`
}
