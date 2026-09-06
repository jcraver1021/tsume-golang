// Package operation holds the outcome of each URL in a download operation, plus
// the folds reducers share. It deliberately depends on nothing else in
// labrador, so anything consuming a finished operation need not pull in the
// downloader.
package operation

// Record is the outcome of one URL.
type Record struct {
	Section  string
	URL      string
	FilePath string
	Success  bool
	Error    error
}
