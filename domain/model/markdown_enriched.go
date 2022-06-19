package model

type LinkEnriched struct {
	//Url  string          `json:"url"`
	Info *GitHubRepoInfo `json:"i"`
}

type MarkdownEnriched struct {
	Links []*LinkEnriched `json:"links"`
}

var EmptyMarkdownEnriched = &MarkdownEnriched{
	Links: []*LinkEnriched{},
}
