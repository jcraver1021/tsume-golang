package mapper

import (
	"html"
	"regexp"
	"strings"
)

var (
	blankRun   = regexp.MustCompile(`\n{3,}`)
	trailingWS = regexp.MustCompile(`[ \t]+\n`)
)

var htmlToText = Mapper{
	Name:     "html-to-text",
	Accepts:  []Kind{KindHTML},
	Produces: KindText,
	Transform: func(payload Payload) (Payload, error) {
		content := htmlScriptBlock.ReplaceAll(payload.Content, nil)
		content = htmlStyleBlock.ReplaceAll(content, nil)
		content = htmlCommentBlock.ReplaceAll(content, nil)
		content = htmlBlockEnd.ReplaceAll(content, []byte("\n"))
		content = htmlAnyTag.ReplaceAll(content, nil)

		text := html.UnescapeString(string(content))
		text = strings.ReplaceAll(text, "\r\n", "\n")
		text = trailingWS.ReplaceAllString(text, "\n")
		text = blankRun.ReplaceAllString(text, "\n\n")
		text = strings.TrimSpace(text) + "\n"

		payload.Content = []byte(text)
		payload.ContentType = "text/plain; charset=utf-8"
		payload.Extension = "txt"
		return payload, nil
	},
}
