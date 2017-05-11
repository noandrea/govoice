package invoice

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
)

var dailyTimeAppItem []struct {
	Activity       string  `json:"activity"`
	DurationString string  `json:"durationString"`
	Percentage     float64 `json:"percentage"`
	Duration       int     `json:"duration"`
}

func scanItemsFromDaily(i *Invoice) {
	dailyExportCommand := fmt.Sprintf(`tell application "Daily" to print json with report "summary" from (date("%s")) to (date("%s"))`, i.Dailytime.DateFrom, i.Dailytime.DateTo)
	//log.Println(dailyExportCommand)
	cmd := exec.Command("/usr/bin/osascript", "-e", dailyExportCommand)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	if err := json.NewDecoder(stdout).Decode(&dailyTimeAppItem); err != nil {
		log.Fatal("Error decoding json: ", err)
	}
	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}

	for _, di := range dailyTimeAppItem {
		//log.Println(di.Activity, " - ", seconds2hours(&di.Duration), " (", di.DurationString, ")")
		for _, pr := range i.Dailytime.Projects {
			if pr.Name == di.Activity {
				i.PushItem(pr.ItemDescription, seconds2hours(&di.Duration), pr.ItemPrice)
			}
		}
	}
}

func seconds2hours(s *int) float64 {
	return (float64(*s) / 60) / 60
}
