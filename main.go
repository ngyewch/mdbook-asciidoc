package main

import (
	"encoding/json"
	"github.com/ngyewch/mdbook-asciidoc/mdbook"
	"github.com/ngyewch/mdbook-asciidoc/renderer"
	"github.com/urfave/cli/v2"
	"io"
	"log"
	"os"
)

var (
	app = &cli.App{
		Name:   "mdbook-asciidoc",
		Usage:  "mdbook-asciidoc",
		Action: doMain,
	}
)

func main() {
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func doMain(cCtx *cli.Context) error {
	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	var renderContext mdbook.RenderContext
	err = json.Unmarshal(b, &renderContext)
	if err != nil {
		return err
	}

	err = renderer.Render(&renderContext, renderer.Config{})
	if err != nil {
		return err
	}

	return nil
}
