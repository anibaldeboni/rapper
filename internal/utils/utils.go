package utils

import (
	"cmp"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"path/filepath"
)

func RandomInt(n int) int64 {
	return Intn(int64(n))
}

// WeightedRandom selects a random value from a map of values based on their weights.
// The values map should contain elements of type T as keys and their corresponding weights as values.
// The function returns the selected value.
func WeightedRandom[T comparable](values map[T]float64) T {
	var chosen T
	var total float64

	if len(values) == 0 {
		return chosen
	}

again:
	for value, weight := range values {
		total += weight
		if RandomFloat64() < weight/total {
			chosen = value
		}
	}
	if IsZero(chosen) {
		goto again
	}
	return chosen
}

func Intn(n int64) int64 {
	nBig, err := rand.Int(rand.Reader, big.NewInt(n))
	if err != nil {
		panic(err)
	}
	return nBig.Int64()
}

func RandomFloat64() float64 {
	return float64(Intn(1<<53)) / (1 << 53)
}

func IsZero[T comparable](val T) bool {
	var zero T
	return zero == val
}

// FindFiles takes a directory path and a list of file patterns as input and returns a list of files that match the patterns in the given directory.
// It also returns a list of any errors encountered during the process.
func FindFiles(dir string, f ...string) ([]string, error) {
	var files []string
	var errs []error
	for _, file := range f {
		found, err := filepath.Glob(filepath.Join(dir, file))
		if err != nil {
			errs = append(errs, err)
		}
		files = append(files, found...)
	}

	if len(errs) > 0 {
		return files, fmt.Errorf("Errors: %w", errors.Join(errs...))
	}

	if len(files) == 0 {
		return files, errors.New("No files found")
	}

	return files, nil
}

// Clamp returns a value clamped between a minimum and maximum value.
// If the value is less than the minimum, it returns the minimum value.
// If the value is greater than the maximum, it returns the maximum value.
// Otherwise, it returns the original value.
func Clamp[T cmp.Ordered](value, minN, maxN T) T {
	if value < minN {
		return minN
	}
	if value > maxN {
		return maxN
	}
	return value
}
