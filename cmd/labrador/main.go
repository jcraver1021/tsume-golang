package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"tsumegolang/internal/labrador"
	"tsumegolang/internal/labrador/config"
	"tsumegolang/internal/labrador/mapper"
	"tsumegolang/internal/labrador/operation"
	"tsumegolang/internal/labrador/reducer"
)

var (
	flagFile        = flag.String("file", "", "file containing a list of URLs to download")
	flagRetryCount  = flag.Int("retry-count", 3, "number of times to retry a failed download")
	flagBackoff     = flag.Int("backoff", 1000, "backoff time in milliseconds between retries")
	flagWorkerCount = flag.Int("worker-count", 1, "number of concurrent workers to use for downloading")
	flagOutputDir   = flag.String("output-dir", "downloads", "base directory for downloaded files")
	flagMap         = flag.String("map", "", "comma-separated mappers applied in order to each download")
	flagReduce      = flag.String("reduce", "markdown-index", "reducer run once over every record, or \"\" for none")
)

func main() {
	flag.Parse()

	if *flagFile == "" {
		log.Fatal("Error: -file flag is required")
	}

	// Both chains resolve before the config is read or a socket is opened, so a
	// bad -map fails in milliseconds rather than after a run's worth of fetches.
	chain, err := mapper.Resolve(strings.Split(*flagMap, ","))
	if err != nil {
		log.Fatalf("Error resolving -map: %v", err)
	}

	fold, hasReducer, err := reducer.Lookup(*flagReduce)
	if err != nil {
		log.Fatalf("Error resolving -reduce: %v", err)
	}

	sections, err := config.Load(*flagFile)
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
		Mappers:     chain,
	})

	downloader.Start()
	defer downloader.Shutdown()

	fmt.Println("Starting downloads...")
	records := downloader.DownloadSections(sections)

	succeeded, _ := operation.CountOutcomes(records)
	fmt.Printf("Downloads completed: %d/%d successful\n", succeeded, len(records))

	if hasReducer {
		summary, err := fold.Reduce(records, *flagOutputDir)
		if err != nil {
			log.Fatalf("Error running reducer %q: %v", fold.Name, err)
		}
		fmt.Println(summary)
	}
}
