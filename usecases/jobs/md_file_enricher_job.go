package jobs

import (
	"context"
	"markdown-enricher/infrastructure/trace"
	"markdown-enricher/pkg/closer"
	"markdown-enricher/usecases/interactors"
	"time"
)

type MdFileEnricherJob struct {
	stopped    chan struct{}
	interactor *interactors.EnricherInteractor
	tp         trace.TracerProvider
}

func MakeMdFileEnricherJob(collection *closer.CloserCollection, tp trace.TracerProvider, interactor *interactors.EnricherInteractor) *MdFileEnricherJob {
	job := &MdFileEnricherJob{
		stopped:    make(chan struct{}),
		interactor: interactor,
		tp:         tp,
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
	tracer := j.tp.Tracer("md-file-parser-job")
	for {
		select {
		case <-j.stopped:
			break
		case mdFile, ok := <-j.interactor.FilesToProcessQueue:
			if !ok {
				break
			}

			ctx, span := tracer.Start(ctx, mdFile.Url)

			timeout, cancelFunc := context.WithTimeout(ctx, 10*time.Minute)
			j.interactor.ForceMarkdown(timeout, mdFile)

			cancelFunc()
			span.End()
		}
	}
}
