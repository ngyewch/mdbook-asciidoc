package main

import (
	"github.com/ngyewch/mdbook-asciidoc/mdbook"
	"github.com/ngyewch/mdbook-asciidoc/renderer"
	"github.com/urfave/cli/v2"
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
	renderContext, err := mdbook.ParseRenderContextFromReader(os.Stdin)
	if err != nil {
		return err
	}

	err = renderer.Render(renderContext, renderer.Config{})
	if err != nil {
		return err
	}

	return nil
}
