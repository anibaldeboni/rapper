package internal

import "path/filepath"

// FindFiles takes a directory path and a list of file patterns as input and returns a list of files that match the patterns in the given directory.
// It also returns a list of any errors encountered during the process.
func FindFiles(dir string, f ...string) ([]string, []error) {
	var files []string
	var errs []error
	for _, file := range f {
		found, err := filepath.Glob(filepath.Join(dir, file))
		if err != nil {
			errs = append(errs, err)
		}
		files = append(files, found...)
	}

	return files, errs
}
