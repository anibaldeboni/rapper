package cli

import (
	"fmt"
	"net/http"
	"rapper/files"
	"rapper/ui"
	"rapper/ui/list"
	"rapper/ui/spinner"
	"rapper/web"
	"sort"

	tea "github.com/charmbracelet/bubbletea"
)

type UpdateMsg struct {
	Message string
	Current int
	Total   int
}

func Run(csvFile files.CSV, hg web.HttpGateway) (err error) {
	errorsCh := make(chan error, len(csvFile.Lines))
	defer close(errorsCh)
	updatesCh := make(chan UpdateMsg, len(csvFile.Lines))
	defer close(updatesCh)

	s := spinner.New()

	go broadcastUpdates(errorsCh, updatesCh, s)
	go execRequests(hg, csvFile, errorsCh, updatesCh)

	if _, err := s.Run(); err != nil {
		return err
	}

	return nil
}

func AskProcessAnotherFile() bool {
	if list.Ask([]string{"Yes", "No"}, ui.Bold("Do you want to process another file?")) == "Yes" {
		return true
	}
	return false
}

func broadcastUpdates(errorsCh <-chan error, updatesCh <-chan UpdateMsg, s *tea.Program) {
	errors := 0
	for {
		select {
		case e := <-errorsCh:
			errors++
			s.Send(spinner.Error(e.Error()))
		case u := <-updatesCh:
			if u.Current == u.Total {
				s.Send(spinner.Done(formatDoneMessage(u.Current, errors)))
				return
			}
			s.Send(spinner.UpdateLabel(fmt.Sprintf("%s %d/%d.", u.Message, u.Current, u.Total)))
		}
	}
}

func execRequests(hg web.HttpGateway, csvFile files.CSV, errorsCh chan<- error, updatesCh chan<- UpdateMsg) {
	total := len(csvFile.Lines)
	progress := 0

	for _, record := range csvFile.Lines {
		response, err := hg.Exec(record)
		if err != nil {
			errorsCh <- fmt.Errorf("%s [%s] %s", ui.IconSkull, ui.Bold("Connection error"), err.Error())
		} else if response.Status != http.StatusOK {
			errorsCh <- fmt.Errorf(formatErrorMsg(record, response.Status))
		}
		progress++
		msg := fmt.Sprintf("Processing %s records", ui.Bold(csvFile.Name))
		updatesCh <- UpdateMsg{Message: msg, Current: progress, Total: total}
	}
}

func formatDoneMessage(recordsCount int, errorsCount int) string {
	var icon string
	var errMsg string
	if errorsCount > 0 {
		icon = ui.IconFireCracker
		errMsg = fmt.Sprintf("%s errors occurred.", ui.Bold(errorsCount))
	} else {
		icon = ui.IconTrophy
		errMsg = fmt.Sprintf("%s.\n", ui.Bold("No errors"))
	}

	return fmt.Sprintf("%s Done! %s records processed. %s", icon, ui.Bold(recordsCount), errMsg)
}

func formatErrorMsg(record map[string]string, status int) string {
	result := ui.IconWarning + "  "
	keys := make([]string, 0, len(record))
	for k := range record {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		result += fmt.Sprintf("%s: %s ", ui.Bold(key), record[key])
	}
	result += fmt.Sprintf("status: %s", ui.Red(status))

	return result
}
