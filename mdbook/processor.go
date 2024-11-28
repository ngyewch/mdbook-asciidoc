package mdbook

import (
	"fmt"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
)

type Handler interface {
	HandleChapter(chapter *Chapter, contentHandler func(walker ast.Walker) error) error
	HandleSeparator(separator *Separator) error
	HandlePartTitle(partTitle *PartTitle) error
}

type processor struct {
	renderContext *RenderContext
	handler       Handler
	md            goldmark.Markdown
}

func Process(renderContext *RenderContext, handler Handler) error {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM, extension.Footnote),
		goldmark.WithParserOptions(),
	)
	p := &processor{
		renderContext: renderContext,
		handler:       handler,
		md:            md,
	}
	return p.process()
}

func (p *processor) process() error {
	for _, section := range p.renderContext.Book.Sections {
		err := p.handleBookItem(section)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *processor) handleBookItem(bookItem *BookItem) error {
	if bookItem.Chapter != nil {
		return p.handleChapter(bookItem.Chapter)
	} else if bookItem.Separator != nil {
		return p.handler.HandleSeparator(bookItem.Separator)
	} else if bookItem.PartTitle != nil {
		return p.handler.HandlePartTitle(bookItem.PartTitle)
	} else {
		return fmt.Errorf("invalid book item")
	}
}

func (p *processor) handleChapter(chapter *Chapter) error {
	err := p.handler.HandleChapter(chapter, func(walker ast.Walker) error {
		sourceBytes := []byte(chapter.Content)
		doc := p.md.Parser().Parse(text.NewReader(sourceBytes))
		return ast.Walk(doc, walker)
	})
	if err != nil {
		return err
	}

	for _, subItem := range chapter.SubItems {
		err := p.handleBookItem(subItem)
		if err != nil {
			return err
		}
	}

	return nil
}
