// Package operation holds the outcome of each URL plus the folds reducers
// share. It depends on nothing else in labrador, so consuming a finished
// operation need not pull in the downloader.
package operation

// Record is the outcome of one URL.
type Record struct {
	Section  string
	URL      string
	FilePath string
	Success  bool
	Error    error
}
