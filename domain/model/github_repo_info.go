package model

import "time"

type GitHubRepoInfo struct {
	Created    time.Time
	Modified   time.Time
	Owner      string `gorm:"primaryKey"`
	Repo       string `gorm:"primaryKey"`
	Stars      int
	Forks      int
	LastCommit time.Time
}
