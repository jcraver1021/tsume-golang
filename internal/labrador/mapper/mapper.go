// Package mapper transforms downloaded bytes before they are written. Mappers
// declare the payload kinds they consume and produce, so a chain can be checked
// for coherence before any network call.
package mapper

import (
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"
)

var (
	ErrUnknownMapper     = errors.New("unknown mapper")
	ErrDuplicateMapper   = errors.New("duplicate mapper")
	ErrUnreachableMapper = errors.New("unreachable mapper")
	ErrTransformFailed   = errors.New("mapper failed")
)

// Kind is the coarse shape of a payload, deliberately coarser than a MIME type:
// it exists to make chains statically checkable.
type Kind string

const (
	KindHTML   Kind = "html"
	KindText   Kind = "text"
	KindJSON   Kind = "json"
	KindXML    Kind = "xml"
	KindBinary Kind = "binary"

	KindSame Kind = "same" // Produces only: the mapper leaves the kind alone
)

// ConcreteKinds is the universe a chain starts from: nothing has been fetched,
// so every kind is possible at position 1.
func ConcreteKinds() []Kind {
	return []Kind{KindBinary, KindHTML, KindJSON, KindText, KindXML}
}

// KindOf classifies a Content-Type header. An absent or unrecognised one is
// binary, so no mapper touches it; guessing wrong would corrupt the download.
func KindOf(contentType string) Kind {
	lower := strings.ToLower(contentType)
	switch {
	case strings.Contains(lower, "html"):
		return KindHTML
	case strings.Contains(lower, "json"):
		return KindJSON
	case strings.Contains(lower, "xml"):
		return KindXML
	case strings.HasPrefix(lower, "text/"):
		return KindText
	default:
		return KindBinary
	}
}

// Payload is the downloaded bytes plus the metadata deciding where they land.
type Payload struct {
	URL         string
	Section     string
	Content     []byte
	ContentType string
	Filename    string // planned name in place of the URL's last segment; empty derives it
	Extension   string // set by a mapper that reshapes content; naming trusts the URL suffix otherwise
}

type Mapper struct {
	Name      string
	Accepts   []Kind // kinds this mapper transforms; others skip it untouched
	Produces  Kind   // kind emitted for an accepted payload, or KindSame
	Transform func(Payload) (Payload, error)
}

func (m Mapper) accepts(kind Kind) bool {
	return slices.Contains(m.Accepts, kind)
}

var registry = map[string]Mapper{
	stripScripts.Name:      stripScripts,
	stripStyles.Name:       stripStyles,
	htmlToText.Name:        htmlToText,
	normalizeNewlines.Name: normalizeNewlines,
}

func Names() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func Lookup(name string) (Mapper, error) {
	mapper, ok := registry[name]
	if !ok {
		return Mapper{}, fmt.Errorf("%w: %q (available: %s)", ErrUnknownMapper, name, strings.Join(Names(), ", "))
	}
	return mapper, nil
}

// Chain is an ordered pipeline applied to every download.
type Chain []Mapper

// Resolve turns flag names into a validated chain, dropping blank entries.
func Resolve(names []string) (Chain, error) {
	chain := make(Chain, 0, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		mapper, err := Lookup(name)
		if err != nil {
			return nil, err
		}
		chain = append(chain, mapper)
	}

	if err := chain.Validate(); err != nil {
		return nil, err
	}
	return chain, nil
}

func (c Chain) Names() []string {
	names := make([]string, len(c))
	for i, mapper := range c {
		names[i] = mapper.Name
	}
	return names
}

// Validate propagates the kinds reaching each position, the way matrix
// dimensions are checked by adjoining them, and rejects any mapper that could
// never fire.
func (c Chain) Validate() error {
	reaching := make(map[Kind]bool, len(ConcreteKinds()))
	for _, kind := range ConcreteKinds() {
		reaching[kind] = true
	}

	consumedBy := make(map[Kind]string)
	seen := make(map[string]int, len(c))

	for i, mapper := range c {
		position := i + 1

		if previous, duplicate := seen[mapper.Name]; duplicate {
			return fmt.Errorf("%w: %q appears at positions %d and %d; running it twice cannot change the result of running it once",
				ErrDuplicateMapper, mapper.Name, previous, position)
		}
		seen[mapper.Name] = position

		accepted := []Kind{}
		for _, kind := range mapper.Accepts {
			if reaching[kind] {
				accepted = append(accepted, kind)
			}
		}

		if len(accepted) == 0 {
			return unreachable(mapper, position, reaching, consumedBy)
		}

		if mapper.Produces == KindSame {
			continue
		}

		for _, kind := range accepted {
			if kind == mapper.Produces {
				continue
			}
			delete(reaching, kind)
			consumedBy[kind] = fmt.Sprintf("%q at position %d", mapper.Name, position)
		}
		reaching[mapper.Produces] = true
	}

	return nil
}

func unreachable(mapper Mapper, position int, reaching map[Kind]bool, consumedBy map[Kind]string) error {
	causes := make([]string, 0, len(mapper.Accepts))
	for _, kind := range mapper.Accepts {
		if culprit, ok := consumedBy[kind]; ok {
			causes = append(causes, fmt.Sprintf("%s is consumed by %s", kind, culprit))
		} else {
			causes = append(causes, fmt.Sprintf("%s never reaches the chain", kind))
		}
	}

	return fmt.Errorf("%w: %q at position %d accepts %s, but %s; kinds reaching position %d: %s",
		ErrUnreachableMapper,
		mapper.Name,
		position,
		joinKinds(mapper.Accepts),
		strings.Join(causes, " and "),
		position,
		joinKinds(sortedKinds(reaching)))
}

func sortedKinds(set map[Kind]bool) []Kind {
	kinds := make([]Kind, 0, len(set))
	for kind := range set {
		kinds = append(kinds, kind)
	}
	slices.Sort(kinds)
	return kinds
}

func joinKinds(kinds []Kind) string {
	parts := make([]string, len(kinds))
	for i, kind := range kinds {
		parts[i] = string(kind)
	}
	return strings.Join(parts, ", ")
}

// Apply runs the chain in order, skipping mappers that do not accept the
// payload's current kind.
func (c Chain) Apply(payload Payload) (Payload, error) {
	for i, mapper := range c {
		if !mapper.accepts(KindOf(payload.ContentType)) {
			continue
		}

		mapped, err := mapper.Transform(payload)
		if err != nil {
			return payload, fmt.Errorf("%w: %q at position %d: %w", ErrTransformFailed, mapper.Name, i+1, err)
		}
		payload = mapped
	}
	return payload, nil
}
