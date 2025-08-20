package renderer

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/ngyewch/mdbook-plugin"
	"github.com/yuin/goldmark/ast"
	extensionAst "github.com/yuin/goldmark/extension/ast"
)

type collector struct {
	renderContext *mdbook.RenderContext
	config        Config
	footnoteIndex int
	footnoteMap   map[footnoteKey]footnoteEntry
}

type footnoteKey struct {
	ChapterId string
	Index     int
}

type footnoteEntry struct {
	Index    int
	Footnote string
}

func getChapterId(chapter *mdbook.Chapter) string {
	if len(chapter.Number) > 0 {
		chapterNumberParts := make([]string, len(chapter.Number))
		for i, chapterNumber := range chapter.Number {
			chapterNumberParts[i] = strconv.Itoa(chapterNumber)
		}
		return strings.Join(chapterNumberParts, "-")
	}
	return ""
}

func (c *collector) HandleChapter(chapter *mdbook.Chapter, contentHandler func(walker ast.Walker) error) error {
	chapterId := getChapterId(chapter)
	err := contentHandler(func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		switch v := node.(type) {
		case *extensionAst.Footnote:
			if entering {
				buf := bytes.NewBuffer(nil)
				mr := &markdownRenderer{
					renderContext: c.renderContext,
					config:        c.config,
					chapter:       chapter,
					sourceBytes:   []byte(chapter.Content),
					w:             buf,
				}
				for child := node.FirstChild(); child != nil; child = child.NextSibling() {
					err := ast.Walk(child, mr.Walk)
					if err != nil {
						return ast.WalkStop, err
					}
				}

				c.footnoteIndex++
				key := footnoteKey{
					ChapterId: chapterId,
					Index:     v.Index,
				}
				value := footnoteEntry{
					Index:    c.footnoteIndex,
					Footnote: buf.String(),
				}
				c.footnoteMap[key] = value

				return ast.WalkSkipChildren, nil
			}
		}

		return ast.WalkContinue, nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *collector) HandleSeparator(separator *mdbook.Separator) error {
	// do nothing
	return nil
}

func (c *collector) HandlePartTitle(title *mdbook.PartTitle) error {
	// do nothing
	return nil
}
