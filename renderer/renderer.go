package renderer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ngyewch/mdbook-plugin"
	"github.com/yuin/goldmark/ast"
)

type renderer struct {
	renderContext *mdbook.RenderContext
	config        Config
	w             io.Writer
	footnoteMap   map[footnoteKey]footnoteEntry
}

type Config struct {
	MinHeadingLevel int `json:"min-heading-level"`
}

func Render(renderContext *mdbook.RenderContext, config Config) error {
	outputDir := renderContext.Destination
	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		return err
	}

	c := &collector{
		renderContext: renderContext,
		config:        config,
		footnoteMap:   make(map[footnoteKey]footnoteEntry),
	}
	err = mdbook.Process(renderContext, c)
	if err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(outputDir, "output.adoc"))
	if err != nil {
		return err
	}

	r := &renderer{
		renderContext: renderContext,
		config:        config,
		w:             f,
		footnoteMap:   c.footnoteMap,
	}

	_, err = fmt.Fprintln(f, ":doctype: book")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(f)
	if err != nil {
		return err
	}

	if renderContext.Config.Book.Title != "" {
		_, err = fmt.Fprintf(f, "= %s\n", renderContext.Config.Book.Title)
		if err != nil {
			return err
		}
	}

	for _, author := range renderContext.Config.Book.Authors {
		_, err = fmt.Fprintf(f, "%s\n", author)
		if err != nil {
			return err
		}
	}

	if renderContext.Config.Book.Description != "" {
		_, err = fmt.Fprintf(f, ":description: %s\n", renderContext.Config.Book.Description)
		if err != nil {
			return err
		}
	}

	_, err = fmt.Fprintln(f, ":toc: left")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(f, ":toclevels: 3")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(f, ":icons: font")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(f, ":table-stripes: even")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(f, ":text-align: left")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(f)
	if err != nil {
		return err
	}

	for _, entry := range r.footnoteMap {
		_, err = fmt.Fprintf(r.w, ":fn-%d: footnote:%d[%s]\n", entry.Index, entry.Index, entry.Footnote)
		if err != nil {
			return err
		}
	}

	err = mdbook.Process(renderContext, r)
	if err != nil {
		return err
	}

	return nil
}

func (r *renderer) HandleChapter(chapter *mdbook.Chapter, contentHandler func(walker ast.Walker) error) error {
	var chapterName string
	var chapterLevel int
	if len(chapter.Number) == 0 {
		chapterName = chapter.Name
		chapterLevel = 2
	} else {
		chapterNumberParts := make([]string, len(chapter.Number))
		for i, chapterNumber := range chapter.Number {
			chapterNumberParts[i] = strconv.Itoa(chapterNumber)
		}
		chapterName = fmt.Sprintf("%s %s", strings.Join(chapterNumberParts, "."), chapter.Name)
		chapterLevel = len(chapter.Number) + 1
	}

	_, err := fmt.Fprintln(r.w)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(r.w, "%s %s\n", strings.Repeat("=", chapterLevel), chapterName)
	if err != nil {
		return err
	}

	mr := &markdownRenderer{
		renderContext: r.renderContext,
		config:        r.config,
		chapter:       chapter,
		sourceBytes:   []byte(chapter.Content),
		w:             r.w,
		footnoteMap:   r.footnoteMap,
	}

	err = contentHandler(mr.Walk)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(r.w)
	if err != nil {
		return err
	}

	return nil
}

func (r *renderer) HandleSeparator(separator *mdbook.Separator) error {
	// do nothing
	return nil
}

func (r *renderer) HandlePartTitle(title *mdbook.PartTitle) error {
	// do nothing
	return nil
}
