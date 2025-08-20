package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ngyewch/mdbook-asciidoc/renderer"
	"github.com/ngyewch/mdbook-plugin"
	"github.com/urfave/cli/v2"
)

var (
	version         string
	commit          string
	commitTimestamp string
)

func main() {
	app := &cli.App{
		Name:    "mdbook-asciidoc",
		Usage:   "mdbook-asciidoc",
		Version: version,
		Action:  doMain,
	}

	cli.VersionPrinter = func(cCtx *cli.Context) {
		var parts []string
		if version != "" {
			parts = append(parts, fmt.Sprintf("version=%s", version))
		}
		if commit != "" {
			parts = append(parts, fmt.Sprintf("commit=%s", commit))
		}
		if commitTimestamp != "" {
			formattedCommitTimestamp := func(commitTimestamp string) string {
				epochSeconds, err := strconv.ParseInt(commitTimestamp, 10, 64)
				if err != nil {
					return ""
				}
				t := time.Unix(epochSeconds, 0)
				return t.Format(time.RFC3339)
			}(commitTimestamp)
			if formattedCommitTimestamp != "" {
				parts = append(parts, fmt.Sprintf("commitTimestamp=%s", formattedCommitTimestamp))
			}
		}
		fmt.Println(strings.Join(parts, " "))
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func doMain(cCtx *cli.Context) error {
	renderContext, err := mdbook.ParseRenderContextFromReader(os.Stdin)
	if err != nil {
		return err
	}

	jsonBytes, err := json.Marshal(renderContext.Config.Output["asciidoc"])
	if err != nil {
		return err
	}

	var config renderer.Config
	err = json.Unmarshal(jsonBytes, &config)
	if err != nil {
		return err
	}

	err = renderer.Render(renderContext, config)
	if err != nil {
		return err
	}

	return nil
}
