package renderer

import (
	"encoding/json"
	"fmt"
	"github.com/ngyewch/mdbook-asciidoc/mdbook"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	extensionAst "github.com/yuin/goldmark/extension/ast"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type renderer struct {
	renderContext *mdbook.RenderContext
	config        Config
	md            goldmark.Markdown
	w             io.Writer
}

type Config struct {
	MinHeadingLevel int `json:"minHeadingLevel"`
}

func Render(renderContext *mdbook.RenderContext, config Config) error {
	outputDir := renderContext.Destination
	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		return err
	}

	{
		f, err := os.Create(filepath.Join(outputDir, "renderContext.json"))
		if err != nil {
			return err
		}
		defer func(f *os.File) {
			_ = f.Close()
		}(f)

		jsonEncoder := json.NewEncoder(f)
		jsonEncoder.SetIndent("", "  ")
		jsonEncoder.SetEscapeHTML(false)
		err = jsonEncoder.Encode(renderContext)
		if err != nil {
			return err
		}
	}

	f, err := os.Create(filepath.Join(outputDir, "output.adoc"))
	if err != nil {
		return err
	}
	r := &renderer{
		renderContext: renderContext,
		config:        config,
		w:             f,
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

	err = mdbook.Process(renderContext, r)
	if err != nil {
		return err
	}

	return nil
}

func (r *renderer) HandleChapter(chapter *mdbook.Chapter, contentHandler func(walker ast.Walker) error) error {
	_, err := fmt.Fprintln(r.w)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(r.w, "<<<")
	if err != nil {
		return err
	}

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

	_, err = fmt.Fprintln(r.w)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(r.w, "%s %s\n", strings.Repeat("=", chapterLevel), chapterName)
	if err != nil {
		return err
	}

	sourceBytes := []byte(chapter.Content)
	nodeLevel := 0
	err = contentHandler(func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			nodeLevel++
		} else {
			nodeLevel--
		}
		switch v := node.(type) {
		case *ast.AutoLink:
			if entering {
				url := string(v.URL(sourceBytes))
				printNode(node, nodeLevel, map[string]any{
					"AutoLinkType": v.AutoLinkType,
					"URL":          url,
				})
				_, err = fmt.Fprintf(r.w, "<%s>", url)
				if err != nil {
					return ast.WalkStop, err
				}
			}

		case *ast.Blockquote:
			if entering {
				printNode(node, nodeLevel, nil)
				_, err = fmt.Fprintln(r.w)
				if err != nil {
					return ast.WalkStop, err
				}
				_, err = fmt.Fprintln(r.w, "____")
				if err != nil {
					return ast.WalkStop, err
				}
			} else {
				_, err = fmt.Fprintln(r.w, "____")
				if err != nil {
					return ast.WalkStop, err
				}
			}

		case *ast.CodeSpan:
			if entering {
				printNode(node, nodeLevel, nil)
				_, err = fmt.Fprint(r.w, "`")
				if err != nil {
					return ast.WalkStop, err
				}
			} else {
				_, err = fmt.Fprint(r.w, "`")
				if err != nil {
					return ast.WalkStop, err
				}
			}

		case *ast.Document:
			if entering {
				printNode(node, nodeLevel, nil)
			}

		case *ast.Emphasis:
			if entering {
				printNode(node, nodeLevel, map[string]any{
					"Level": v.Level,
				})
				switch v.Level {
				case 1:
					_, err = fmt.Fprint(r.w, "_")
					if err != nil {
						return ast.WalkStop, err
					}
				case 2:
					_, err = fmt.Fprint(r.w, "*")
					if err != nil {
						return ast.WalkStop, err
					}
				default:
					return ast.WalkStop, fmt.Errorf("unknown emphasis level: %d", v.Level)
				}
			} else {
				switch v.Level {
				case 1:
					_, err = fmt.Fprint(r.w, "_")
					if err != nil {
						return ast.WalkStop, err
					}
				case 2:
					_, err = fmt.Fprint(r.w, "*")
					if err != nil {
						return ast.WalkStop, err
					}
				default:
					return ast.WalkStop, fmt.Errorf("unknown emphasis level: %d", v.Level)
				}
			}

		case *ast.FencedCodeBlock:
			if entering {
				language := string(v.Language(sourceBytes))
				printNode(node, nodeLevel, map[string]any{
					"Language": language,
				})
				_, err = fmt.Fprintln(r.w)
				if err != nil {
					return ast.WalkStop, err
				}
				if language != "" {
					_, err = fmt.Fprintf(r.w, "[source,%s]\n", language)
					if err != nil {
						return ast.WalkStop, err
					}
				} else {
					_, err = fmt.Fprintln(r.w, "[source]")
					if err != nil {
						return ast.WalkStop, err
					}
				}
				_, err = fmt.Fprintln(r.w, "----")
				if err != nil {
					return ast.WalkStop, err
				}

				lines := v.Lines()
				for i := 0; i < lines.Len(); i++ {
					line := lines.At(i)
					l := string(line.Value(sourceBytes))
					_, err = fmt.Fprint(r.w, l)
					if err != nil {
						return ast.WalkStop, err
					}
				}
				_, err = fmt.Fprintln(r.w, "----")
				if err != nil {
					return ast.WalkStop, err
				}
			}

		case *ast.HTMLBlock:
			if entering {
				_, err = fmt.Fprint(r.w, "pass:[")
				if err != nil {
					return ast.WalkStop, err
				}
				lines := v.Lines()
				for i := 0; i < lines.Len(); i++ {
					line := lines.At(i)
					l := string(line.Value(sourceBytes))
					_, err = fmt.Fprint(r.w, l)
					if err != nil {
						return ast.WalkStop, err
					}
				}
			} else {
				_, err = fmt.Fprint(r.w, "]")
				if err != nil {
					return ast.WalkStop, err
				}
			}

		case *ast.Heading:
			if (r.config.MinHeadingLevel > 0) && v.Level < r.config.MinHeadingLevel {
				return ast.WalkSkipChildren, nil
			}
			if entering {
				printNode(node, nodeLevel, map[string]any{
					"Level": v.Level,
				})

				_, err = fmt.Fprintln(r.w)
				if err != nil {
					return ast.WalkStop, err
				}
				_, err = fmt.Fprintln(r.w, "[discrete]")
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
				_, err = fmt.Fprintf(r.w, "%s ", strings.Repeat("=", asciidocHeadingLevel))
				if err != nil {
					return ast.WalkStop, err
				}
			} else {
				_, err = fmt.Fprintln(r.w)
				if err != nil {
					return ast.WalkStop, err
				}
			}

		case *ast.Image:
			if entering {
				destination := string(v.Destination)
				title := string(v.Title)
				printNode(node, nodeLevel, map[string]any{
					"Destination": destination,
					"Title":       title,
				})
				u, err := url.Parse(destination)
				if err != nil {
					return ast.WalkStop, err
				}
				if (u.Scheme == "") || (u.Scheme == "file") {
					src := r.renderContext.Config.Book.Src
					if src == "" {
						src = "src"
					}
					baseDirectory := filepath.Join(r.renderContext.Root, src)
					sourcePath := filepath.Join(baseDirectory, filepath.Dir(chapter.SourcePath), u.Path)
					relPath := filepath.Join(filepath.Dir(chapter.SourcePath), u.Path)
					destinationPath := filepath.Join(r.renderContext.Destination, relPath)
					fmt.Println(sourcePath, destinationPath)
					err = doCopyFile(sourcePath, destinationPath)
					if err != nil {
						return ast.WalkStop, err
					}
					destination = relPath
				}
				_, err = fmt.Fprintf(r.w, "image:%s[", destination)
				if err != nil {
					return ast.WalkStop, err
				}
				if title != "" {
					_, err = fmt.Fprintf(r.w, "title=%s,", title)
					if err != nil {
						return ast.WalkStop, err
					}
				}
			} else {
				_, err = fmt.Fprint(r.w, "]")
				if err != nil {
					return ast.WalkStop, err
				}
			}

		case *ast.Link:
			if entering {
				title := string(v.Title)
				destination := string(v.Destination)
				printNode(node, nodeLevel, map[string]any{
					"Title":       title,
					"Destination": destination,
				})
				_, err = fmt.Fprintf(r.w, "link:%s[", destination)
				if err != nil {
					return ast.WalkStop, err
				}
			} else {
				_, err = fmt.Fprint(r.w, "]")
				if err != nil {
					return ast.WalkStop, err
				}
			}

		case *ast.List:
			if entering {
				printNode(node, nodeLevel, map[string]any{
					"Start":   v.Start,
					"Marker":  string(v.Marker),
					"IsTight": v.IsTight,
				})
				_, err = fmt.Fprintln(r.w)
				if err != nil {
					return ast.WalkStop, err
				}
				if v.IsOrdered() && (v.Start != 1) {
					_, err = fmt.Fprintf(r.w, "[start=%d]\n", v.Start)
					if err != nil {
						return ast.WalkStop, err
					}
				} else {
					_, err = fmt.Fprint(r.w, "[disc]\n")
					if err != nil {
						return ast.WalkStop, err
					}
				}
			}

		case *ast.ListItem:
			if entering {
				printNode(node, nodeLevel, map[string]any{
					"Offset": v.Offset,
				})
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
				_, err = fmt.Fprintf(r.w, "%s ", strings.Repeat(marker, level))
				if err != nil {
					return ast.WalkStop, err
				}
			} else {
				_, err = fmt.Fprintln(r.w)
				if err != nil {
					return ast.WalkStop, err
				}
			}

		case *ast.Paragraph:
			if entering {
				printNode(node, nodeLevel, nil)
				switch node.Parent().(type) {
				case *ast.ListItem, *extensionAst.Footnote:
					// do nothing
				default:
					_, err = fmt.Fprintln(r.w)
					if err != nil {
						return ast.WalkStop, err
					}
				}
			} else {
				switch node.Parent().(type) {
				case *extensionAst.Footnote:
					// do nothing
				default:
					_, err = fmt.Fprintln(r.w)
					if err != nil {
						return ast.WalkStop, err
					}
				}
			}

		case *ast.Text:
			if entering {
				value := string(v.Value(sourceBytes))
				printNode(node, nodeLevel, map[string]any{
					"Value": value,
				})
				_, err = fmt.Fprint(r.w, value)
				if err != nil {
					return ast.WalkStop, err
				}
			}

		case *ast.TextBlock:
			if entering {
				printNode(node, nodeLevel, nil)
			}

		case *ast.ThematicBreak:
			if entering {
				printNode(node, nodeLevel, nil)
				_, err = fmt.Fprintln(r.w)
				if err != nil {
					return ast.WalkStop, err
				}
				_, err = fmt.Fprintln(r.w, "'''")
				if err != nil {
					return ast.WalkStop, err
				}
			}

		case *extensionAst.Footnote: // TODO collect footnotes
			if entering {
				ref := string(v.Ref)
				printNode(node, nodeLevel, map[string]any{
					"Ref":   ref,
					"Index": v.Index,
				})
				_, err = fmt.Fprintln(r.w)
				if err != nil {
					return ast.WalkStop, err
				}
				_, err = fmt.Fprintf(r.w, ":fn-%d: footnote:%d[", v.Index, v.Index)
				if err != nil {
					return ast.WalkStop, err
				}
			} else {
				_, err = fmt.Fprintln(r.w, "]")
				if err != nil {
					return ast.WalkStop, err
				}
			}

		case *extensionAst.FootnoteBacklink:
			// not supported
			if entering {
				printNode(node, nodeLevel, map[string]any{
					"Index":    v.Index,
					"RefCount": v.RefCount,
					"RefIndex": v.RefIndex,
				})
				return ast.WalkSkipChildren, nil
			}

		case *extensionAst.FootnoteLink: // TODO collect footnotes
			if entering {
				printNode(node, nodeLevel, map[string]any{
					"Index":    v.Index,
					"RefCount": v.RefCount,
					"RefIndex": v.RefIndex,
				})
				_, err = fmt.Fprintf(r.w, "{fn-%d}", v.Index)
				if err != nil {
					return ast.WalkStop, err
				}
			}

		case *extensionAst.FootnoteList:
			if entering {
				printNode(node, nodeLevel, map[string]any{
					"Count": v.Count,
				})
			}

		case *extensionAst.Strikethrough:
			if entering {
				printNode(node, nodeLevel, nil)
				_, err = fmt.Fprint(r.w, "[.line-through]#")
				if err != nil {
					return ast.WalkStop, err
				}
			} else {
				_, err = fmt.Fprint(r.w, "#")
				if err != nil {
					return ast.WalkStop, err
				}
			}

		case *extensionAst.Table:
			if entering {
				printNode(node, nodeLevel, map[string]any{
					"Alignments": v.Alignments,
				})
				_, err = fmt.Fprintln(r.w)
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
				_, err = fmt.Fprintf(r.w, "[%s]\n", strings.Join(tableSpecs, ","))
				if err != nil {
					return ast.WalkStop, err
				}
				_, err = fmt.Fprintln(r.w, "|===")
				if err != nil {
					return ast.WalkStop, err
				}
			} else {
				_, err = fmt.Fprintln(r.w, "|===")
				if err != nil {
					return ast.WalkStop, err
				}
			}

		case *extensionAst.TableCell:
			if entering {
				printNode(node, nodeLevel, nil)
				_, err = fmt.Fprint(r.w, "|")
				if err != nil {
					return ast.WalkStop, err
				}
			} else {
				_, err = fmt.Fprint(r.w, " ")
				if err != nil {
					return ast.WalkStop, err
				}
			}

		case *extensionAst.TableHeader:
			if entering {
				printNode(node, nodeLevel, nil)
			} else {
				_, err = fmt.Fprintln(r.w)
				if err != nil {
					return ast.WalkStop, err
				}
			}

		case *extensionAst.TableRow:
			if entering {
				printNode(node, nodeLevel, nil)
			} else {
				_, err = fmt.Fprintln(r.w)
				if err != nil {
					return ast.WalkStop, err
				}
			}

		default:
			if entering {
				doPrintNode(node, nodeLevel, false, nil)
			}
		}

		return ast.WalkContinue, nil
	})
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

func printNode(node ast.Node, nodeLevel int, attrs map[string]any) {
	doPrintNode(node, nodeLevel, true, attrs)
}

func doPrintNode(node ast.Node, nodeLevel int, handled bool, attrs map[string]any) {
	fmt.Print(strings.Repeat(" ", nodeLevel))
	fmt.Print("- ")
	if handled {
		fmt.Printf("[%v]", node.Kind())
	} else {
		fmt.Printf("[! %v]", node.Kind())
	}
	if attrs != nil {
		fmt.Printf(" %v", attrs)
	}
	fmt.Println()
}

func doCopyFile(src string, dst string) error {
	r, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func(r *os.File) {
		_ = r.Close()
	}(r)

	err = os.MkdirAll(filepath.Dir(dst), 0755)
	if err != nil {
		return err
	}

	w, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func(w *os.File) {
		_ = w.Close()
	}(w)

	_, err = io.Copy(w, r)
	if err != nil {
		return err
	}

	return nil
}
