package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"

	"tsumegolang/internal/labrador"
)

var (
	flagFile        = flag.String("file", "", "file containing a list of URLs to download")
	flagRetryCount  = flag.Int("retry-count", 3, "number of times to retry a failed download")
	flagBackoff     = flag.Int("backoff", 1000, "backoff time in milliseconds between retries")
	flagWorkerCount = flag.Int("worker-count", 1, "number of concurrent workers to use for downloading")
	flagOutputDir   = flag.String("output-dir", "downloads", "base directory for downloaded files")
)

func main() {
	flag.Parse()

	if *flagFile == "" {
		log.Fatal("Error: -file flag is required")
	}

	sections, err := labrador.ParseSectionsFromYAML(*flagFile)
	if err != nil {
		log.Fatalf("Error parsing YAML file: %v", err)
	}

	if len(sections) == 0 {
		log.Fatal("Error: no valid sections found in YAML file")
	}

	downloader := labrador.NewMultiDownloader(labrador.MultiDownloaderSettings{
		RetryCount:  *flagRetryCount,
		BackoffMs:   *flagBackoff,
		WorkerCount: *flagWorkerCount,
		OutputDir:   *flagOutputDir,
	})

	downloader.Start()
	defer downloader.Shutdown()

	fmt.Println("Starting downloads...")
	records := downloader.DownloadSections(sections)

	indexPath := filepath.Join(*flagOutputDir, "index.md")
	err = labrador.GenerateMarkdownIndex(records, indexPath)
	if err != nil {
		log.Fatalf("Error generating markdown index: %v", err)
	}

	successCount := 0
	for _, record := range records {
		if record.Success {
			successCount++
		}
	}

	fmt.Printf("Downloads completed: %d/%d successful\n", successCount, len(records))
	fmt.Printf("Index generated at: %s\n", indexPath)
}
