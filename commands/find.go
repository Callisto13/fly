package commands

import (
	"errors"
	"os"
	"strconv"

	"github.com/concourse/fly/commands/internal/displayhelpers"
	"github.com/concourse/fly/rc"
	"github.com/concourse/fly/ui"
	"github.com/fatih/color"
)

type FindCommand struct {
	Json  bool `long:"json" description:"Print command result as JSON"`
	Build int  `short:"b" long:"build-id" description:"Build ID to get worker information for"`
}

func (command *FindCommand) Execute([]string) error {
	if command.Build == 0 {
		return errors.New("Must specify --build-id")
	}

	target, err := rc.LoadTarget(Fly.Target, Fly.Verbose)
	if err != nil {
		return err
	}

	err = target.Validate()
	if err != nil {
		return err
	}

	containers, err := target.Team().ListContainers(map[string]string{})
	if err != nil {
		return err
	}

	//TODO: this crap
	if command.Json {
		var cut map[string]string
		for _, c := range containers {
			if c.BuildID == command.Build {
				cut = map[string]string{
					"build_id":    strconv.Itoa(c.BuildID),
					"worker_name": c.WorkerName,
				}
			}
		}
		err = displayhelpers.JsonPrint(cut)
		if err != nil {
			return err
		}
		return nil
	}

	table := ui.Table{
		Headers: ui.TableRow{
			{Contents: "build id", Color: color.New(color.Bold)},
			{Contents: "worker", Color: color.New(color.Bold)},
		},
	}

	var row ui.TableRow
	for _, c := range containers {
		if c.BuildID == command.Build {
			row = ui.TableRow{
				{Contents: strconv.Itoa(c.BuildID)},
				{Contents: c.WorkerName},
			}
			table.Data = append(table.Data, row)
		}
	}

	if len(table.Data) == 0 {
		row = ui.TableRow{
			{Contents: strconv.Itoa(command.Build)},
			{Contents: "no worker found"},
		}
		table.Data = append(table.Data, row)
	}

	return table.Render(os.Stdout, Fly.PrintTableHeaders)
}
