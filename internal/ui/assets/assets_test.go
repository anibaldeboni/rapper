package assets_test

import (
	"testing"

	"github.com/anibaldeboni/rapper/internal/ui/assets"
	"github.com/stretchr/testify/assert"
)

func TestLoadAllFiglets(t *testing.T) {
	figlets, err := assets.LoadAllFiglets()
	assert.NoError(t, err)

	// Assert that the figlets map is not empty
	assert.NotEmpty(t, figlets)

	// Assert that each figlet has a non-empty value
	for name, font := range figlets {
		assert.NotEmpty(t, font, "Font for figlet '%s' is empty", name)
	}

	// Assert that the figlet names are correct
	expectedFigletNames := []string{"bloody", "crawford", "crazy", "epic", "fraktur", "ghoulish", "larry3d", "merlin1", "nancyj", "poison", "rozzo", "script", "small"}
	for _, name := range expectedFigletNames {
		_, ok := figlets[name]
		assert.True(t, ok, "Figlet '%s' is missing", name)
	}

	// Assert that there are no extra figlets
	for name := range figlets {
		assert.Contains(t, expectedFigletNames, name, "Unexpected figlet '%s'", name)
	}
}
