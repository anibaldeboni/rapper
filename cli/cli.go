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
	"sync"
)

type UpdateMsg struct {
	Type    string
	Message string
	Current int
	Total   int
}

func Run(csvFile files.CSV, hg web.HttpGateway, s spinner.Spinner) error {
	ch := make(chan spinner.UpdateUI, len(csvFile.Lines))
	var wg = &sync.WaitGroup{}

	wg.Add(2)
	go broadcastUpdates(ch, s, wg)
	go execRequests(hg, csvFile, ch, wg)

	go func(ch chan<- spinner.UpdateUI, s spinner.Spinner) {
		if _, err := s.Run(); err != nil {
			close(ch)
			Exit(err)
		}
	}(ch, s)

	wg.Wait()
	return nil
}

func AskProcessAnotherFile() bool {
	options := []list.Option[bool]{
		{
			Title: "Yes",
			Value: true,
		},
		{
			Title: "No",
			Value: false,
		},
	}
	if answer, err := list.Ask(options, ui.Bold("Do you want to process another file?")); err == nil {
		return answer
	}

	return false
}

func broadcastUpdates(ch <-chan spinner.UpdateUI, s spinner.Spinner, wg *sync.WaitGroup) {
	defer wg.Done()
	for u := range ch {
		s.Update(u)
	}
}

func execRequests(hg web.HttpGateway, csvFile files.CSV, ch chan<- spinner.UpdateUI, wg *sync.WaitGroup) {
	defer close(ch)
	defer wg.Done()

	total := len(csvFile.Lines)
	errs := 0

	for i, record := range csvFile.Lines {
		response, err := hg.Exec(record)
		if err != nil {
			errs++
			ch <- spinner.UpdateUI{
				Type:    spinner.Error,
				Message: fmt.Sprintf("[%s] %s", ui.Bold("Connection error"), err.Error()),
			}
		} else if response.Status != http.StatusOK {
			errs++
			ch <- spinner.UpdateUI{
				Type:    spinner.Error,
				Message: formatErrorMsg(record, response.Status),
			}
		}
		ch <- spinner.UpdateUI{
			Type:    spinner.Update,
			Message: fmt.Sprintf("Processing %s records %d/%d", ui.Bold(csvFile.Name), i+1, total),
		}
	}
	ch <- spinner.UpdateUI{Type: spinner.Done, Message: formatDoneMessage(csvFile.Name, total, errs)}
}

func formatDoneMessage(fileName string, recordsCount int, errorsCount int) string {
	var icon string
	var errMsg string
	if errorsCount > 0 {
		icon = ui.IconFireCracker
		errMsg = fmt.Sprintf("%s errors occurred.", ui.Bold(errorsCount))
	} else {
		icon = ui.IconTrophy
		errMsg = fmt.Sprintf("%s.\n", ui.Bold("No errors"))
	}

	return fmt.Sprintf("%s Done! %s %s records processed. %s", icon, ui.Bold(recordsCount), ui.Green(fileName), errMsg)
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
