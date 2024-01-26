package cli

import (
	"fmt"
	"os"
	"rapper/ui"
)

func Exit(message any, arg ...any) {
	switch message.(type) {
	case string:
		fmt.Println(ui.QuitTextStyle.Render(fmt.Sprintf(message.(string), arg...)))
		os.Exit(0)
	case error:
		fmt.Println(ui.QuitTextStyle.Render(fmt.Sprintf(message.(error).Error()+"\n", arg...)))
		os.Exit(1)
	default:
		fmt.Println(ui.QuitTextStyle.Render(fmt.Sprintf("%v\n", message)))
		os.Exit(1)
	}
}
