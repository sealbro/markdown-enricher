package model

import "time"

type GitHubRepoInfo struct {
	Created    time.Time `json:"-"`
	Modified   time.Time `json:"-"`
	Owner      string    `json:"o" gorm:"primaryKey"`
	Repo       string    `json:"r" gorm:"primaryKey"`
	Stars      int       `json:"s"`
	Forks      int       `json:"f"`
	LastCommit time.Time `json:"lc"`
}
