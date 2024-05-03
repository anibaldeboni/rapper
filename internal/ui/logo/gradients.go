package logo

import (
	"github.com/anibaldeboni/rapper/internal/utils"
	"github.com/mazznoer/colorgrad"
)

var (
	Warm               = colorgrad.Warm()
	Cool               = colorgrad.Cool()
	Spectral           = colorgrad.Spectral()
	GoldToTurquoise, _ = colorgrad.NewGradient().
				HtmlColors("gold", "hotpink", "darkturquoise").
				Build()
	PinkToGreen, _ = colorgrad.NewGradient().
			HtmlColors("deeppink", "gold", "seagreen").
			Build()
	PinkToPurple, _ = colorgrad.NewGradient().
			HtmlColors("#C41189", "#00BFFF").
			Build()
)

func randomColorGradient() colorgrad.Gradient {

	grads := []colorgrad.Gradient{
		PinkToPurple,
		PinkToGreen,
		GoldToTurquoise,
		Warm,
		Cool,
		Spectral,
	}

	return grads[utils.RandomInt(len(grads))]
}
