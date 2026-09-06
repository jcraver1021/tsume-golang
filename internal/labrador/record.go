package labrador

// DownloadRecord is the outcome of one URL: what a reducer folds over.
type DownloadRecord struct {
	Section  string
	URL      string
	FilePath string
	Success  bool
	Error    error
}
