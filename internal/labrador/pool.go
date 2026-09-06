package labrador

import (
	"errors"
	"fmt"
	"time"

	"tsumegolang/internal/labrador/config"
	"tsumegolang/internal/labrador/fetch"
	"tsumegolang/internal/labrador/mapper"
	"tsumegolang/internal/labrador/operation"
	"tsumegolang/internal/labrador/store"
	"tsumegolang/pkg/concurrency"
)

const (
	defaultWorkerCount = 1
)

var (
	ErrDownloadFailed      = errors.New("download failed")
	ErrJobSubmissionFailed = errors.New("job submission failed")
	ErrTimeout             = errors.New("timeout waiting for result")
)

type downloadJob struct {
	URL     string
	Section string
}

type MultiDownloader struct {
	workerPool *concurrency.WorkerPool[downloadJob, operation.Record]
	outputDir  string
}

type MultiDownloaderSettings struct {
	RetryCount  int
	BackoffMs   int
	WorkerCount int
	OutputDir   string
	Mappers     mapper.Chain
}

func NewMultiDownloader(settings MultiDownloaderSettings) *MultiDownloader {
	numWorkers := settings.WorkerCount
	if numWorkers < 1 {
		numWorkers = defaultWorkerCount
	}

	outputDir := settings.OutputDir
	if outputDir == "" {
		outputDir = "."
	}

	fetchOpts := []fetch.Option{fetch.WithIdleConnsPerHost(numWorkers)}
	if settings.RetryCount > 0 {
		fetchOpts = append(fetchOpts, fetch.WithRetryCount(settings.RetryCount))
	}
	if settings.BackoffMs > 0 {
		fetchOpts = append(fetchOpts, fetch.WithBackoff(settings.BackoffMs))
	}
	fetcher := fetch.New(fetchOpts...)

	job := func(dj downloadJob) concurrency.JobResult[downloadJob, operation.Record] {
		result, err := fetcher.Get(dj.URL)

		record := operation.Record{
			Section: dj.Section,
			URL:     dj.URL,
			Success: false,
		}

		if err != nil {
			record.Error = fmt.Errorf("%w: %w", ErrDownloadFailed, err)
			return concurrency.JobResult[downloadJob, operation.Record]{
				Input:  dj,
				Output: record,
				Err:    record.Error,
				Status: concurrency.StatusError,
			}
		}

		payload, err := settings.Mappers.Apply(mapper.Payload{
			URL:         dj.URL,
			Section:     dj.Section,
			Content:     result.Content,
			ContentType: result.ContentType,
		})
		if err != nil {
			record.Error = err
			return concurrency.JobResult[downloadJob, operation.Record]{
				Input:  dj,
				Output: record,
				Err:    err,
				Status: concurrency.StatusError,
			}
		}

		filePath, err := store.Write(payload, outputDir)
		if err != nil {
			record.Error = err
			return concurrency.JobResult[downloadJob, operation.Record]{
				Input:  dj,
				Output: record,
				Err:    err,
				Status: concurrency.StatusError,
			}
		}

		record.Success = true
		record.FilePath = filePath

		return concurrency.JobResult[downloadJob, operation.Record]{
			Input:  dj,
			Output: record,
			Status: concurrency.StatusSuccess,
		}
	}

	return &MultiDownloader{
		workerPool: concurrency.NewWorkerPool(job, numWorkers),
		outputDir:  outputDir,
	}
}

func (md *MultiDownloader) Start() {
	md.workerPool.Start()
}

func (md *MultiDownloader) DownloadSections(sections []config.Section) []operation.Record {
	var allJobs []downloadJob
	for _, section := range sections {
		for _, url := range section.URLs {
			allJobs = append(allJobs, downloadJob{
				URL:     url,
				Section: section.Name,
			})
		}
	}

	results := make([]concurrency.JobResult[downloadJob, operation.Record], len(allJobs))
	resultChans := make([]chan concurrency.JobResult[downloadJob, operation.Record], len(allJobs))

	for i, job := range allJobs {
		resultCh, err := md.workerPool.Submit(job)
		if err != nil {
			record := operation.Record{
				Section: job.Section,
				URL:     job.URL,
				Success: false,
				Error:   fmt.Errorf("%w: %w", ErrJobSubmissionFailed, err),
			}
			results[i] = concurrency.JobResult[downloadJob, operation.Record]{
				Input:  job,
				Output: record,
				Err:    record.Error,
				Status: concurrency.StatusError,
			}
			continue
		}
		resultChans[i] = resultCh
	}

	for i, resultCh := range resultChans {
		if resultCh == nil {
			continue
		}
		select {
		case result := <-resultCh:
			results[i] = result
		case <-time.After(10 * time.Minute):
			record := operation.Record{
				Section: allJobs[i].Section,
				URL:     allJobs[i].URL,
				Success: false,
				Error:   fmt.Errorf("%w: %s", ErrTimeout, allJobs[i].URL),
			}
			results[i] = concurrency.JobResult[downloadJob, operation.Record]{
				Input:  allJobs[i],
				Output: record,
				Err:    record.Error,
				Status: concurrency.StatusError,
			}
		}
	}

	records := make([]operation.Record, len(results))
	for i, result := range results {
		records[i] = result.Output
	}

	return records
}

func (md *MultiDownloader) Shutdown() {
	md.workerPool.Shutdown()
}
