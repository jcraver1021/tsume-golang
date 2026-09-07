package mapper

import (
	"regexp"
)

// Patterns shared by the HTML-consuming mappers.
var (
	htmlScriptBlock  = regexp.MustCompile(`(?is)<script\b[^>]*>.*?</script\s*>`)
	htmlStyleBlock   = regexp.MustCompile(`(?is)<style\b[^>]*>.*?</style\s*>`)
	htmlCommentBlock = regexp.MustCompile(`(?s)<!--.*?-->`)
	htmlBlockEnd     = regexp.MustCompile(`(?i)<br\s*/?>|</(p|div|li|tr|h[1-6]|blockquote|section|article)\s*>`)
	htmlAnyTag       = regexp.MustCompile(`(?s)<[^>]*>`)
)
