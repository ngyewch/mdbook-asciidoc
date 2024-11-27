package renderer

import (
	"encoding/json"
	"fmt"
	"github.com/ngyewch/mdbook-asciidoc/mdbook"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extensionAst "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	headingPointSizes = []float64{27, 22.5, 18, 13.5, 10.5, 9}
)

type session struct {
	renderContext *mdbook.RenderContext
	md            goldmark.Markdown
	w             io.Writer
}

func Render(renderContext *mdbook.RenderContext) error {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(),
	)

	outputDir := filepath.Join(renderContext.Destination, "asciidoc")
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
	s := &session{
		renderContext: renderContext,
		md:            md,
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

	_, err = fmt.Fprintln(f)
	if err != nil {
		return err
	}

	err = s.render()
	if err != nil {
		return err
	}

	return nil
}

func (s *session) render() error {
	for _, section := range s.renderContext.Book.Sections {
		err := s.handleBookItem(section)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *session) handleBookItem(bookItem *mdbook.BookItem) error {
	if bookItem.Chapter != nil {
		return s.handleChapter(bookItem.Chapter)
	} else if bookItem.Separator != nil {
		return s.handleSeparator(bookItem.Separator)
	} else if bookItem.PartTitle != nil {
		return s.handlePartTitle(bookItem.PartTitle)
	} else {
		return fmt.Errorf("invalid book item")
	}
}

func printNode(node ast.Node, nodeLevel int, attrs map[string]any) {
	fmt.Printf("%s- [%T / %v]", strings.Repeat(" ", nodeLevel), node, node.Kind())
	if attrs != nil {
		fmt.Printf(" %v", attrs)
	}
	fmt.Println()
}

func (s *session) handleChapter(chapter *mdbook.Chapter) error {
	_, err := fmt.Fprintln(s.w)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(s.w, "<<<")
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

	_, err = fmt.Fprintln(s.w)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(s.w, "%s %s\n", strings.Repeat("=", chapterLevel), chapterName)
	if err != nil {
		return err
	}

	sourceBytes := []byte(chapter.Content)
	doc := s.md.Parser().Parse(text.NewReader(sourceBytes))
	nodeLevel := 0
	err = ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			nodeLevel++
		} else {
			nodeLevel--
		}
		switch v := node.(type) {
		case *ast.Document:
			if entering {
				printNode(node, nodeLevel, nil)
			}

		case *ast.Heading:
			if v.Level < 2 {
				return ast.WalkSkipChildren, nil
			}
			if entering {
				printNode(node, nodeLevel, map[string]any{
					"Level": v.Level,
				})

				if (v.Level < 1) || (v.Level > 5) {
					return ast.WalkStop, fmt.Errorf("unsupported heading level: %d", v.Level)
				}

				_, err = fmt.Fprintln(s.w)
				if err != nil {
					return ast.WalkStop, err
				}
				_, err = fmt.Fprintln(s.w, "[discrete]")
				if err != nil {
					return ast.WalkStop, err
				}
				_, err = fmt.Fprintf(s.w, "%s ", strings.Repeat("=", v.Level+1))
				if err != nil {
					return ast.WalkStop, err
				}
			} else {
				_, err = fmt.Fprintln(s.w)
				if err != nil {
					return ast.WalkStop, err
				}
			}

		case *ast.Paragraph:
			if entering {
				printNode(node, nodeLevel, nil)
				_, err = fmt.Fprintln(s.w)
				if err != nil {
					return ast.WalkStop, err
				}
			} else {
				_, err = fmt.Fprintln(s.w)
				if err != nil {
					return ast.WalkStop, err
				}
			}

		case *ast.Text:
			if entering {
				value := string(v.Value(sourceBytes))
				printNode(node, nodeLevel, map[string]any{
					"Value": value,
				})
				_, err = fmt.Fprint(s.w, value)
				if err != nil {
					return ast.WalkStop, err
				}
			}

		case *ast.Emphasis:
			if entering {
				fmt.Printf("%s- [%T / %v] Level: %d\n", strings.Repeat(" ", nodeLevel), node, node.Kind(), v.Level)
				switch v.Level {
				case 1:
					/*
						s.pdfHelper.PushDrawSettings(
							WithItalic(),
						)
					*/
				case 2:
					/*
						s.pdfHelper.PushDrawSettings(
							WithBold(),
						)
					*/
				default:
					return ast.WalkStop, fmt.Errorf("unknown emphasis level: %d", v.Level)
				}
			} else {
				/*
					s.pdfHelper.PopDrawSettings()
				*/
			}

		case *ast.CodeSpan:
			if entering {
				fmt.Printf("%s- [%T / %v]\n", strings.Repeat(" ", nodeLevel), node, node.Kind())
				/*
					s.pdfHelper.PushDrawSettings(
						WithTypefaceClass(CodeTypefaceClass),
					)
					s.pdfHelper.pdf.SetFillColor(255, 255, 0)
					s.pdfHelper.pdf.SetFillSpotColor("code", 100)
				*/
			} else {
				/*
					s.pdfHelper.PopDrawSettings()
				*/
			}

		case *ast.Link:
			if entering {
				title := string(v.Title)
				destination := string(v.Destination)
				printNode(node, nodeLevel, map[string]any{
					"Title":       title,
					"Destination": destination,
				})
				_, err = fmt.Fprintf(s.w, "link:%s[", destination)
				if err != nil {
					return ast.WalkStop, err
				}
			} else {
				_, err = fmt.Fprint(s.w, "]")
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
				_, err = fmt.Fprintln(s.w)
				if err != nil {
					return ast.WalkStop, err
				}
				_, err = fmt.Fprintf(s.w, "[start=%d]\n", v.Start)
				if err != nil {
					return ast.WalkStop, err
				}
			}

		case *ast.ListItem:
			if entering {
				printNode(node, nodeLevel, map[string]any{
					"Offset": v.Offset,
				})
				parent := v.Parent().(*ast.List)
				_, err = fmt.Fprintf(s.w, "%c ", parent.Marker)
				if err != nil {
					return ast.WalkStop, err
				}
			} else {
				_, err = fmt.Fprintln(s.w)
				if err != nil {
					return ast.WalkStop, err
				}
			}

		case *extensionAst.Table:
			if entering {
				/*
					rows := v.ChildCount()
					cols := 0
					for row := v.FirstChild(); row != nil; row = row.NextSibling() {
						cols = max(cols, row.ChildCount())
					}
					colWidths := make([]float64, cols)
					for row := v.FirstChild(); row != nil; row = row.NextSibling() {
						for i, col := 0, row.FirstChild(); col != nil; i, col = i+1, col.NextSibling() {
							txt := string(col.Text(sourceBytes))
							w := s.pdf.GetStringWidth(txt)
							colWidths[i] = max(colWidths[i], w)
						}
					}
					fmt.Printf("%s- [%T / %v] rows: %d, cols: %d, colWidths: %v\n", strings.Repeat(" ", nodeLevel), node, node.Kind(), rows, cols, colWidths)
				*/
			}

		case *extensionAst.TableHeader:
			if entering {
				fmt.Printf("%s- [%T / %v]\n", strings.Repeat(" ", nodeLevel), node, node.Kind())
			}

		case *extensionAst.TableRow:
			if entering {
				fmt.Printf("%s- [%T / %v]\n", strings.Repeat(" ", nodeLevel), node, node.Kind())
			}

		case *extensionAst.TableCell:
			if entering {
				fmt.Printf("%s- [%T / %v]\n", strings.Repeat(" ", nodeLevel), node, node.Kind())
			}

		default:
			if entering {
				txt := string(v.Text(sourceBytes))
				fmt.Printf("%s- !!! [%T / %v] %s\n", strings.Repeat(" ", nodeLevel), node, node.Kind(), txt)
			}
		}

		return ast.WalkContinue, nil
	})
	if err != nil {
		return err
	}

	for _, subItem := range chapter.SubItems {
		err := s.handleBookItem(subItem)
		if err != nil {
			return err
		}
	}
	// TODO
	return nil
}

func (s *session) handleSeparator(separator *mdbook.Separator) error {
	// TODO
	return nil
}

func (s *session) handlePartTitle(title *mdbook.PartTitle) error {
	// TODO
	return nil
}
