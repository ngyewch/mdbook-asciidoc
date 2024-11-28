package renderer

import (
	"fmt"
	"github.com/ngyewch/mdbook-asciidoc/mdbook"
	"github.com/yuin/goldmark/ast"
	extensionAst "github.com/yuin/goldmark/extension/ast"
	"io"
	"log"
	"net/url"
	"path/filepath"
	"strings"
)

type markdownRenderer struct {
	renderContext *mdbook.RenderContext // TODO remove
	config        Config
	chapter       *mdbook.Chapter // TODO remove
	sourceBytes   []byte
	w             io.Writer
	footnoteMap   map[footnoteKey]footnoteEntry // TODO refactor
}

func (mr *markdownRenderer) Walk(node ast.Node, entering bool) (ast.WalkStatus, error) {
	switch v := node.(type) {
	case *ast.AutoLink:
		if entering {
			u := string(v.URL(mr.sourceBytes))
			_, err := fmt.Fprintf(mr.w, "<%s>", u)
			if err != nil {
				return ast.WalkStop, err
			}
		}

	case *ast.Blockquote:
		if entering {
			_, err := fmt.Fprintln(mr.w)
			if err != nil {
				return ast.WalkStop, err
			}
			_, err = fmt.Fprintln(mr.w, "____")
			if err != nil {
				return ast.WalkStop, err
			}
		} else {
			_, err := fmt.Fprintln(mr.w, "____")
			if err != nil {
				return ast.WalkStop, err
			}
		}

	case *ast.CodeSpan:
		if entering {
			_, err := fmt.Fprint(mr.w, "`")
			if err != nil {
				return ast.WalkStop, err
			}
		} else {
			_, err := fmt.Fprint(mr.w, "`")
			if err != nil {
				return ast.WalkStop, err
			}
		}

	case *ast.Emphasis:
		if entering {
			switch v.Level {
			case 1:
				_, err := fmt.Fprint(mr.w, "_")
				if err != nil {
					return ast.WalkStop, err
				}
			case 2:
				_, err := fmt.Fprint(mr.w, "*")
				if err != nil {
					return ast.WalkStop, err
				}
			default:
				return ast.WalkStop, fmt.Errorf("unknown emphasis level: %d", v.Level)
			}
		} else {
			switch v.Level {
			case 1:
				_, err := fmt.Fprint(mr.w, "_")
				if err != nil {
					return ast.WalkStop, err
				}
			case 2:
				_, err := fmt.Fprint(mr.w, "*")
				if err != nil {
					return ast.WalkStop, err
				}
			default:
				return ast.WalkStop, fmt.Errorf("unknown emphasis level: %d", v.Level)
			}
		}

	case *ast.FencedCodeBlock:
		if entering {
			language := string(v.Language(mr.sourceBytes))
			_, err := fmt.Fprintln(mr.w)
			if err != nil {
				return ast.WalkStop, err
			}
			if language != "" {
				_, err = fmt.Fprintf(mr.w, "[source,%s]\n", language)
				if err != nil {
					return ast.WalkStop, err
				}
			} else {
				_, err = fmt.Fprintln(mr.w, "[source]")
				if err != nil {
					return ast.WalkStop, err
				}
			}
			_, err = fmt.Fprintln(mr.w, "----")
			if err != nil {
				return ast.WalkStop, err
			}

			lines := v.Lines()
			for i := 0; i < lines.Len(); i++ {
				line := lines.At(i)
				l := string(line.Value(mr.sourceBytes))
				_, err = fmt.Fprint(mr.w, l)
				if err != nil {
					return ast.WalkStop, err
				}
			}
			_, err = fmt.Fprintln(mr.w, "----")
			if err != nil {
				return ast.WalkStop, err
			}
		}

	case *ast.HTMLBlock:
		if entering {
			_, err := fmt.Fprint(mr.w, "pass:[")
			if err != nil {
				return ast.WalkStop, err
			}
			lines := v.Lines()
			for i := 0; i < lines.Len(); i++ {
				line := lines.At(i)
				l := string(line.Value(mr.sourceBytes))
				_, err = fmt.Fprint(mr.w, l)
				if err != nil {
					return ast.WalkStop, err
				}
			}
		} else {
			_, err := fmt.Fprint(mr.w, "]")
			if err != nil {
				return ast.WalkStop, err
			}
		}

	case *ast.Heading:
		if (mr.config.MinHeadingLevel > 0) && v.Level < mr.config.MinHeadingLevel {
			return ast.WalkSkipChildren, nil
		}
		if entering {
			_, err := fmt.Fprintln(mr.w)
			if err != nil {
				return ast.WalkStop, err
			}
			_, err = fmt.Fprintln(mr.w, "[discrete]")
			if err != nil {
				return ast.WalkStop, err
			}
			asciidocHeadingLevel := v.Level
			if asciidocHeadingLevel < 1 {
				asciidocHeadingLevel = 1
				log.Printf("unsupported heading level %d, changed to %d", v.Level+1, asciidocHeadingLevel)
			} else if asciidocHeadingLevel > 6 {
				asciidocHeadingLevel = 6
				log.Printf("unsupported heading level %d, changed to %d", v.Level+1, asciidocHeadingLevel)
			}
			_, err = fmt.Fprintf(mr.w, "%s ", strings.Repeat("=", asciidocHeadingLevel))
			if err != nil {
				return ast.WalkStop, err
			}
		} else {
			_, err := fmt.Fprintln(mr.w)
			if err != nil {
				return ast.WalkStop, err
			}
		}

	case *ast.Image:
		if entering {
			destination := string(v.Destination)
			title := string(v.Title)
			u, err := url.Parse(destination)
			if err != nil {
				return ast.WalkStop, err
			}
			if (u.Scheme == "") || (u.Scheme == "file") {
				src := mr.renderContext.Config.Book.Src
				if src == "" {
					src = "src"
				}
				baseDirectory := filepath.Join(mr.renderContext.Root, src)
				sourcePath := filepath.Join(baseDirectory, filepath.Dir(mr.chapter.SourcePath), u.Path)
				relPath := filepath.Join(filepath.Dir(mr.chapter.SourcePath), u.Path)
				destinationPath := filepath.Join(mr.renderContext.Destination, relPath)
				fmt.Println(sourcePath, destinationPath)
				err = copyFile(sourcePath, destinationPath)
				if err != nil {
					return ast.WalkStop, err
				}
				destination = relPath
			}
			_, err = fmt.Fprintf(mr.w, "image:%s[", destination)
			if err != nil {
				return ast.WalkStop, err
			}
			if title != "" {
				_, err = fmt.Fprintf(mr.w, "title=%s,", title)
				if err != nil {
					return ast.WalkStop, err
				}
			}
		} else {
			_, err := fmt.Fprint(mr.w, "]")
			if err != nil {
				return ast.WalkStop, err
			}
		}

	case *ast.Link:
		if entering {
			destination := string(v.Destination)
			_, err := fmt.Fprintf(mr.w, "link:%s[", destination)
			if err != nil {
				return ast.WalkStop, err
			}
		} else {
			_, err := fmt.Fprint(mr.w, "]")
			if err != nil {
				return ast.WalkStop, err
			}
		}

	case *ast.List:
		if entering {
			_, err := fmt.Fprintln(mr.w)
			if err != nil {
				return ast.WalkStop, err
			}
			if v.IsOrdered() && (v.Start != 1) {
				_, err = fmt.Fprintf(mr.w, "[start=%d]\n", v.Start)
				if err != nil {
					return ast.WalkStop, err
				}
			} else {
				_, err = fmt.Fprint(mr.w, "[disc]\n")
				if err != nil {
					return ast.WalkStop, err
				}
			}
		}

	case *ast.ListItem:
		if entering {
			level := 0
			n := node
			for n != nil {
				switch n.(type) {
				case *ast.List:
					level++
				}
				n = n.Parent()
			}
			parent := v.Parent().(*ast.List)
			marker := "*"
			if parent.IsOrdered() {
				marker = "."
			}
			_, err := fmt.Fprintf(mr.w, "%s ", strings.Repeat(marker, level))
			if err != nil {
				return ast.WalkStop, err
			}
		} else {
			_, err := fmt.Fprintln(mr.w)
			if err != nil {
				return ast.WalkStop, err
			}
		}

	case *ast.Paragraph:
		if entering {
			switch node.Parent().(type) {
			case *ast.ListItem, *extensionAst.Footnote:
				// do nothing
			default:
				_, err := fmt.Fprintln(mr.w)
				if err != nil {
					return ast.WalkStop, err
				}
			}
		} else {
			switch node.Parent().(type) {
			case *extensionAst.Footnote:
				// do nothing
			default:
				_, err := fmt.Fprintln(mr.w)
				if err != nil {
					return ast.WalkStop, err
				}
			}
		}

	case *ast.Text:
		if entering {
			value := string(v.Value(mr.sourceBytes))
			_, err := fmt.Fprint(mr.w, value)
			if err != nil {
				return ast.WalkStop, err
			}
		}

	case *ast.ThematicBreak:
		if entering {
			_, err := fmt.Fprintln(mr.w)
			if err != nil {
				return ast.WalkStop, err
			}
			_, err = fmt.Fprintln(mr.w, "'''")
			if err != nil {
				return ast.WalkStop, err
			}
		}

	case *extensionAst.Footnote:
		if entering {
			return ast.WalkSkipChildren, nil
		}

	case *extensionAst.FootnoteBacklink:
		// not supported
		if entering {
			return ast.WalkSkipChildren, nil
		}

	case *extensionAst.FootnoteLink:
		if entering {
			if mr.footnoteMap != nil {
				key := footnoteKey{
					ChapterId: getChapterId(mr.chapter),
					Index:     v.Index,
				}
				value, ok := mr.footnoteMap[key]
				if ok {
					_, err := fmt.Fprintf(mr.w, "{fn-%d}", value.Index)
					if err != nil {
						return ast.WalkStop, err
					}
				}
			}
		}

	case *extensionAst.Strikethrough:
		if entering {
			_, err := fmt.Fprint(mr.w, "[.line-through]#")
			if err != nil {
				return ast.WalkStop, err
			}
		} else {
			_, err := fmt.Fprint(mr.w, "#")
			if err != nil {
				return ast.WalkStop, err
			}
		}

	case *extensionAst.Table:
		if entering {
			_, err := fmt.Fprintln(mr.w)
			if err != nil {
				return ast.WalkStop, err
			}
			var cols []string
			for _, alignment := range v.Alignments {
				switch alignment {
				case extensionAst.AlignLeft:
					cols = append(cols, "<")
				case extensionAst.AlignRight:
					cols = append(cols, ">")
				case extensionAst.AlignCenter:
					cols = append(cols, "^")
				case extensionAst.AlignNone:
					cols = append(cols, "")
				default:
					cols = append(cols, "")
				}
			}
			var options []string
			_, hasTableHeader := v.FirstChild().(*extensionAst.TableHeader)
			if hasTableHeader {
				options = append(options, "header")
			}
			var tableSpecs []string
			if len(cols) > 0 {
				tableSpecs = append(tableSpecs, fmt.Sprintf("cols=\"%s\"", strings.Join(cols, ",")))
			}
			if len(options) > 0 {
				tableSpecs = append(tableSpecs, fmt.Sprintf("options=\"%s\"", strings.Join(options, ",")))
			}
			_, err = fmt.Fprintf(mr.w, "[%s]\n", strings.Join(tableSpecs, ","))
			if err != nil {
				return ast.WalkStop, err
			}
			_, err = fmt.Fprintln(mr.w, "|===")
			if err != nil {
				return ast.WalkStop, err
			}
		} else {
			_, err := fmt.Fprintln(mr.w, "|===")
			if err != nil {
				return ast.WalkStop, err
			}
		}

	case *extensionAst.TableCell:
		if entering {
			_, err := fmt.Fprint(mr.w, "|")
			if err != nil {
				return ast.WalkStop, err
			}
		} else {
			_, err := fmt.Fprint(mr.w, " ")
			if err != nil {
				return ast.WalkStop, err
			}
		}

	case *extensionAst.TableHeader:
		if entering {
			// do nothing
		} else {
			_, err := fmt.Fprintln(mr.w)
			if err != nil {
				return ast.WalkStop, err
			}
		}

	case *extensionAst.TableRow:
		if entering {
			// do nothing
		} else {
			_, err := fmt.Fprintln(mr.w)
			if err != nil {
				return ast.WalkStop, err
			}
		}
	}

	return ast.WalkContinue, nil
}
