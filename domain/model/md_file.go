package model

import "time"

type JobStatus int

const (
	Ready JobStatus = iota
	Process
	Done
)

type MdFile struct {
	Created  time.Time `json:"-"`
	Modified time.Time `json:"-"`
	Url      string    `gorm:"primaryKey"`
	Status   JobStatus
}
