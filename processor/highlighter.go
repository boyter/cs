package processor

import (
	"bytes"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
)

type InputLanguage struct {
	FileName        string `json:"fileName"`
	Style           string `json:"style"`
	Content         string `json:"content"`
	WithLineNumbers bool   `json:"withLineNumbers"`
}

type OutputLanguage struct {
	Css  string `json:"css"`
	Html string `json:"html"`
}

func Highlight(inputLanguage InputLanguage) (OutputLanguage, error) {

	lexer := lexers.Match(inputLanguage.FileName)
	if lexer == nil {
		lexer = lexers.Analyse(inputLanguage.Content)

		if lexer == nil {
			lexer = lexers.Fallback
		}
	}

	style := styles.Get(inputLanguage.Style)
	if style == nil {
		style = styles.Fallback
	}

	// Parse the content
	iterator, err := lexer.Tokenise(nil, inputLanguage.Content)
	if err != nil {
		return OutputLanguage{}, err
	}

	var cssBytes bytes.Buffer
	var htmlBytes bytes.Buffer

	formatter := html.New(html.WithLineNumbers(), html.WithClasses())
	if formatter.WriteCSS(&cssBytes, style) != nil {
		return OutputLanguage{}, err
	}

	if formatter.Format(&htmlBytes, style, iterator) != nil {
		return OutputLanguage{}, err
	}

	return OutputLanguage{
		Css:  cssBytes.String(),
		Html: htmlBytes.String(),
	}, nil
}
