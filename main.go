package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/ngyewch/mdbook-asciidoc/renderer"
	"github.com/ngyewch/mdbook-plugin"
	"github.com/urfave/cli/v3"
)

var (
	version string
)

func main() {
	app := &cli.Command{
		Name:    "mdbook-asciidoc",
		Usage:   "mdbook-asciidoc",
		Version: version,
		Action:  doMain,
	}

	err := app.Run(context.Background(), os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func doMain(ctx context.Context, cmd *cli.Command) error {
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
