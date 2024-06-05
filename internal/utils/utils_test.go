package utils_test

import (
	"testing"

	"github.com/anibaldeboni/rapper/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestWeightedRandom(t *testing.T) {
	t.Run("Should return a value from the map based on weights", func(t *testing.T) {
		values := map[string]float64{
			"apple":  0.5,
			"banana": 0.3,
			"orange": 0.2,
		}

		result := utils.WeightedRandom(values)

		assert.Contains(t, values, result)
	})

	t.Run("Should return the only value when there is only one in the map", func(t *testing.T) {
		values := map[string]float64{
			"apple": 1.0,
		}

		result := utils.WeightedRandom(values)

		assert.Equal(t, "apple", result)
	})

	t.Run("Should return an empty value when the map is empty", func(t *testing.T) {
		values := map[string]float64{}

		result := utils.WeightedRandom(values)

		assert.Empty(t, result)
	})
}
