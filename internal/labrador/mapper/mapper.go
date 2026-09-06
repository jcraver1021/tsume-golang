// Package mapper defines the per-download transformations labrador applies
// between fetching bytes and writing them to disk. Every mapper declares the
// payload kinds it consumes and produces, so a chain can be checked for
// coherence before a single network call is made.
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

// Kind is the coarse shape of a payload's content. It is deliberately coarser
// than a MIME type: it exists to make chains statically checkable, not to
// describe a file precisely.
type Kind string

const (
	KindHTML   Kind = "html"
	KindText   Kind = "text"
	KindJSON   Kind = "json"
	KindXML    Kind = "xml"
	KindBinary Kind = "binary"

	// KindSame is only valid as a Produces value; it declares that a mapper
	// hands back whatever kind it was given.
	KindSame Kind = "same"
)

// ConcreteKinds is the universe a chain starts from. Validation runs before any
// URL is fetched, so every kind must be assumed possible at position 1.
func ConcreteKinds() []Kind {
	return []Kind{KindBinary, KindHTML, KindJSON, KindText, KindXML}
}

// KindOf classifies a Content-Type header. An absent or unrecognised header is
// binary, which means no mapper touches it — the conservative choice, since
// guessing wrong would corrupt the download.
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

// Payload is the unit a Mapper transforms: the downloaded bytes plus the
// metadata that decides where they land. A mapper that changes the shape of the
// content must set Extension, because file naming otherwise trusts the URL
// suffix over ContentType.
type Payload struct {
	URL         string
	Section     string
	Content     []byte
	ContentType string
	Extension   string
}

type Mapper struct {
	Name string
	// Accepts is the set of kinds this mapper transforms. Payloads of any other
	// kind skip it untouched.
	Accepts []Kind
	// Produces is the kind emitted for an accepted payload, or KindSame when the
	// mapper leaves the kind alone.
	Produces  Kind
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

// Chain is an ordered mapper pipeline applied to every download.
type Chain []Mapper

// Resolve turns flag names into a validated chain. Blank entries are dropped so
// an empty -map flag yields an empty chain rather than an error.
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

// Validate propagates the set of kinds that can reach each position, the way
// matrix dimensions are checked by adjoining them, and rejects a chain in which
// some mapper could never fire.
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

// Apply runs the chain in order, skipping any mapper whose Accepts set does not
// cover the payload's current kind.
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
