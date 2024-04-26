package logo

import (
	"github.com/anibaldeboni/rapper/internal/utils"
	"github.com/mazznoer/colorgrad"
)

var (
	warm               = colorgrad.Warm()
	cool               = colorgrad.Cool()
	spectral           = colorgrad.Spectral()
	goldToTurquoise, _ = colorgrad.NewGradient().
				HtmlColors("gold", "hotpink", "darkturquoise").
				Build()
	pinkToGreen, _ = colorgrad.NewGradient().
			HtmlColors("deeppink", "gold", "seagreen").
			Build()
	pinkToPurple, _ = colorgrad.NewGradient().
			HtmlColors("#C41189", "#00BFFF").
			Build()
)

func randomColorGradient() colorgrad.Gradient {

	grads := []colorgrad.Gradient{
		pinkToPurple,
		pinkToGreen,
		goldToTurquoise,
		warm,
		cool,
		spectral,
	}

	return grads[utils.RandomInt(len(grads))]
}
