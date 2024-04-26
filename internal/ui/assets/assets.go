package assets

import (
	"embed"

	"github.com/anibaldeboni/rapper/internal/utils"
)

//go:embed files/*
var logoFiles embed.FS

type LogoName string

const (
	MainLogo LogoName = "logo.txt"
	Crawford LogoName = "crawford.txt"
	Fraktur  LogoName = "fraktur.txt"
	Ghoulish LogoName = "ghoulish.txt"
	Larry3D  LogoName = "larry3d.txt"
	Merlin1  LogoName = "merlin1.txt"
	NancyJ   LogoName = "nancyj.txt"
	Poison   LogoName = "poison.txt"
	Rozzo    LogoName = "rozzo.txt"
	Script   LogoName = "script.txt"
	Small    LogoName = "small.txt"
)

func Logo() string {
	logo, _ := logoFiles.ReadFile("files/logo.txt")

	return string(logo)
}

func LogoByName(name LogoName) string {
	logo, _ := logoFiles.ReadFile("files/" + string(name))

	return string(logo)
}

// RandomLogo returns a randomly selected logo from the "files" directory.
func RandomLogo() string {
	files, _ := logoFiles.ReadDir("files")
	var logos []string
	for _, file := range files {
		if !file.IsDir() {
			logo, _ := logoFiles.ReadFile("files/" + file.Name())
			logos = append(logos, string(logo))
		}
	}

	return logos[utils.RandomInt(len(logos))]
}
