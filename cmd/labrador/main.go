package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"tsumegolang/internal/labrador"
	"tsumegolang/internal/labrador/config"
	"tsumegolang/internal/labrador/mapper"
	"tsumegolang/internal/labrador/operation"
	"tsumegolang/internal/labrador/reducer"
)

var (
	flagFile        = flag.String("file", "", "YAML file describing the sections and URLs to download")
	flagFrom        = flag.String("from", "", "manifest from an earlier run; re-applies reducers without downloading")
	flagRetryCount  = flag.Int("retry-count", 3, "number of times to retry a failed download")
	flagBackoff     = flag.Int("backoff", 1000, "backoff time in milliseconds between retries")
	flagWorkerCount = flag.Int("worker-count", 1, "number of concurrent workers to use for downloading")
	flagOutputDir   = flag.String("output-dir", "downloads", "base directory for downloaded files")
	flagMap         = flag.String("map", "", "comma-separated mappers applied in order to each download")
	flagReduce      = flag.String("reduce", "markdown-index", "comma-separated reducers, each folding every record into its own artifact")
)

// downloadOnlyFlags have no meaning when -from re-applies reducers to a
// finished operation.
var downloadOnlyFlags = []string{"file", "retry-count", "backoff", "worker-count", "map"}

func main() {
	flag.Parse()
	provided := providedFlags()

	if err := checkFlags(provided); err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Reducers resolve before anything is read or fetched, so an unrunnable set
	// costs milliseconds rather than a whole operation.
	reducers, err := reducer.Resolve(strings.Split(*flagReduce, ","))
	if err != nil {
		log.Fatalf("Error resolving -reduce: %v", err)
	}

	var records []operation.Record
	options := reducer.Options{OutputDir: *flagOutputDir}

	if *flagFrom != "" {
		// Artifacts belong beside the manifest they were derived from unless
		// the caller says otherwise.
		if !provided["output-dir"] {
			options.OutputDir = filepath.Dir(*flagFrom)
		}
		options.Source = *flagFrom

		records, err = reducer.LoadJSONManifest(*flagFrom)
		if err != nil {
			log.Fatalf("Error loading manifest: %v", err)
		}
		fmt.Printf("Loaded %d records from %s\n", len(records), *flagFrom)
	} else {
		records = download()
	}

	report, err := reducer.Run(reducers, records, options)
	for _, summary := range report.Summaries {
		fmt.Println(summary)
	}
	for _, skip := range report.Skips {
		fmt.Println(skip)
	}
	if err != nil {
		log.Fatalf("Reducers failed:\n%v", err)
	}
}

func providedFlags() map[string]bool {
	provided := map[string]bool{}
	flag.Visit(func(f *flag.Flag) { provided[f.Name] = true })
	return provided
}

func checkFlags(provided map[string]bool) error {
	switch {
	case *flagFile == "" && *flagFrom == "":
		return fmt.Errorf("one of -file or -from is required")
	case *flagFile != "" && *flagFrom != "":
		return fmt.Errorf("-file and -from are mutually exclusive; -from re-applies reducers to a finished operation")
	}

	if *flagFrom == "" {
		return nil
	}
	for _, name := range downloadOnlyFlags {
		if provided[name] {
			return fmt.Errorf("-%s has no meaning with -from, which downloads nothing", name)
		}
	}
	return nil
}

func download() []operation.Record {
	chain, err := mapper.Resolve(strings.Split(*flagMap, ","))
	if err != nil {
		log.Fatalf("Error resolving -map: %v", err)
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

	return records
}
