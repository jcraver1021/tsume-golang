package mapper

import (
	"strings"
)

var normalizeNewlines = Mapper{
	Name:     "normalize-newlines",
	Accepts:  []Kind{KindText, KindJSON, KindXML},
	Produces: KindSame,
	Transform: func(payload Payload) (Payload, error) {
		content := strings.ReplaceAll(string(payload.Content), "\r\n", "\n")
		payload.Content = []byte(strings.ReplaceAll(content, "\r", "\n"))
		return payload, nil
	},
}
