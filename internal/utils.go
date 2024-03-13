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

// Clamp returns a value clamped between a minimum and maximum value.
// If the value is less than the minimum, it returns the minimum value.
// If the value is greater than the maximum, it returns the maximum value.
// Otherwise, it returns the original value.
func Clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
