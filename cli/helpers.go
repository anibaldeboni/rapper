package cli

import (
	"fmt"
	"os"
	"rapper/ui"
)

func ExitOnError(message string, arg ...any) {
	fmt.Println(ui.QuitTextStyle.Render(fmt.Sprintf(message+"\n", arg...)))
	os.Exit(1)
}
