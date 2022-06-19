package jobs

import (
	"context"
	"markdown-enricher/pkg/closer"
	"markdown-enricher/usecases/interactors"
	"time"
)

type MdFileEnricherJob struct {
	stopped    chan struct{}
	interactor *interactors.EnricherInteractor
}

func MakeMdFileEnricherJob(collection *closer.CloserCollection, interactor *interactors.EnricherInteractor) *MdFileEnricherJob {
	job := &MdFileEnricherJob{
		stopped:    make(chan struct{}),
		interactor: interactor,
	}

	collection.Add(job)

	return job
}

func (j *MdFileEnricherJob) Close(context.Context) error {
	j.stopped <- struct{}{}

	return nil
}

func (j *MdFileEnricherJob) Start() {
	ctx := context.Background()
	for {
		select {
		case <-j.stopped:
			break
		case mdFile := <-j.interactor.FilesToProcessQueue:
			timeout, cancelFunc := context.WithTimeout(ctx, 10*time.Minute)
			j.interactor.ForceMarkdown(timeout, mdFile)

			cancelFunc()
		}
	}
}
