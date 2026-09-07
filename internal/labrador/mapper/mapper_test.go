package mapper_test

import (
	"errors"
	"strings"
	"testing"

	. "tsumegolang/internal/labrador/mapper"
)

// applyOne runs a single registered mapper's transform directly, bypassing the
// chain's kind guard so a mapper's own behaviour can be tested in isolation.
func applyOne(t *testing.T, name string, payload Payload) Payload {
	t.Helper()

	mapper, err := Lookup(name)
	if err != nil {
		t.Fatalf("Lookup(%q) = %v", name, err)
	}

	got, err := mapper.Transform(payload)
	if err != nil {
		t.Fatalf("%s.Transform() = %v", name, err)
	}
	return got
}

func applyChain(t *testing.T, names []string, payload Payload) Payload {
	t.Helper()

	chain, err := Resolve(names)
	if err != nil {
		t.Fatalf("Resolve(%v) = %v", names, err)
	}

	got, err := chain.Apply(payload)
	if err != nil {
		t.Fatalf("Apply() = %v", err)
	}
	return got
}

func TestRegistryInvariants(t *testing.T) {
	valid := map[Kind]bool{}
	for _, kind := range ConcreteKinds() {
		valid[kind] = true
	}

	for _, name := range Names() {
		t.Run(name, func(t *testing.T) {
			mapper, err := Lookup(name)
			if err != nil {
				t.Fatalf("Lookup(%q) = %v", name, err)
			}

			if mapper.Name != name {
				t.Errorf("Name = %q, want %q — registry key must match", mapper.Name, name)
			}
			if len(mapper.Accepts) == 0 {
				t.Error("Accepts is empty; a mapper that accepts nothing can never fire")
			}
			for _, kind := range mapper.Accepts {
				if !valid[kind] {
					t.Errorf("Accepts contains %q, which is not a concrete kind", kind)
				}
			}
			if mapper.Produces != KindSame && !valid[mapper.Produces] {
				t.Errorf("Produces = %q, want a concrete kind or KindSame", mapper.Produces)
			}
			if mapper.Transform == nil {
				t.Error("Transform is nil")
			}
		})
	}
}

func TestKindOf(t *testing.T) {
	testCases := []struct {
		name        string
		contentType string
		want        Kind
	}{
		{name: "html", contentType: "text/html", want: KindHTML},
		{name: "html with charset", contentType: "text/html; charset=utf-8", want: KindHTML},
		{name: "xhtml is html", contentType: "application/xhtml+xml", want: KindHTML},
		{name: "json", contentType: "application/json", want: KindJSON},
		{name: "xml", contentType: "application/xml", want: KindXML},
		{name: "text xml is xml not text", contentType: "text/xml", want: KindXML},
		{name: "plain text", contentType: "text/plain", want: KindText},
		{name: "markdown", contentType: "text/markdown", want: KindText},
		{name: "image", contentType: "image/png", want: KindBinary},
		{name: "pdf", contentType: "application/pdf", want: KindBinary},
		{name: "absent header is binary", contentType: "", want: KindBinary},
		{name: "case insensitive", contentType: "TEXT/HTML", want: KindHTML},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := KindOf(tc.contentType); got != tc.want {
				t.Errorf("KindOf(%q) = %q, want %q", tc.contentType, got, tc.want)
			}
		})
	}
}

func TestResolve(t *testing.T) {
	testCases := []struct {
		name            string
		names           []string
		want            []string
		wantErr         error
		wantErrContains []string
	}{
		{
			name:  "empty flag yields an empty chain",
			names: []string{""},
			want:  []string{},
		},
		{
			name:  "order is preserved",
			names: []string{"strip-styles", "strip-scripts", "html-to-text"},
			want:  []string{"strip-styles", "strip-scripts", "html-to-text"},
		},
		{
			name:  "surrounding whitespace is tolerated",
			names: []string{" strip-scripts ", "  ", "html-to-text"},
			want:  []string{"strip-scripts", "html-to-text"},
		},
		{
			name:  "text mapper after a conversion is fine",
			names: []string{"html-to-text", "normalize-newlines"},
			want:  []string{"html-to-text", "normalize-newlines"},
		},
		{
			name:            "unknown name is rejected",
			names:           []string{"strip-scripts", "transmogrify"},
			wantErr:         ErrUnknownMapper,
			wantErrContains: append([]string{`"transmogrify"`}, Names()...),
		},
		{
			name:            "repeating a mapper is rejected",
			names:           []string{"strip-scripts", "strip-styles", "strip-scripts"},
			wantErr:         ErrDuplicateMapper,
			wantErrContains: []string{`"strip-scripts"`, "positions 1 and 3"},
		},
		{
			name:    "html mapper after html is consumed is rejected",
			names:   []string{"html-to-text", "strip-scripts"},
			wantErr: ErrUnreachableMapper,
			wantErrContains: []string{
				`"strip-scripts" at position 2`,
				"accepts html",
				`html is consumed by "html-to-text" at position 1`,
				"kinds reaching position 2: binary, json, text, xml",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Resolve(tc.names)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("err = %v, want %v", err, tc.wantErr)
				}
				if got != nil {
					t.Error("chain should be nil when resolution fails")
				}
				for _, want := range tc.wantErrContains {
					if !strings.Contains(err.Error(), want) {
						t.Errorf("error %q does not contain %q", err, want)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("Resolve() = %v", err)
			}

			names := got.Names()
			if len(names) != len(tc.want) {
				t.Fatalf("names = %v, want %v", names, tc.want)
			}
			for i, name := range names {
				if name != tc.want[i] {
					t.Errorf("names[%d] = %q, want %q", i, name, tc.want[i])
				}
			}
		})
	}
}

func TestValidateAcceptsCoherentChains(t *testing.T) {
	testCases := []struct {
		name  string
		names []string
	}{
		{name: "kind-preserving mapper before a conversion", names: []string{"normalize-newlines", "html-to-text"}},
		{name: "kind-preserving mapper after a conversion", names: []string{"html-to-text", "normalize-newlines"}},
		{name: "two html mappers then the conversion", names: []string{"strip-scripts", "strip-styles", "html-to-text"}},
		{
			name:  "every mapper, ordered so each can still fire",
			names: []string{"strip-scripts", "strip-styles", "html-to-text", "normalize-newlines"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Resolve(tc.names); err != nil {
				t.Errorf("Resolve(%v) = %v, want it to validate", tc.names, err)
			}
		})
	}
}

func TestApply(t *testing.T) {
	testCases := []struct {
		name          string
		chain         []string
		payload       Payload
		wantContent   string
		wantExtension string
	}{
		{
			name:          "the chain runs in order",
			chain:         []string{"strip-scripts", "html-to-text"},
			payload:       Payload{ContentType: "text/html", Content: []byte(`<style>p{}</style><script>secret</script><p>a &lt; b</p>`)},
			wantContent:   "a < b\n",
			wantExtension: "txt",
		},
		{
			name:        "a mapper that does not accept the kind is skipped",
			chain:       []string{"strip-scripts", "normalize-newlines"},
			payload:     Payload{ContentType: "text/html", Content: []byte("<p>a</p>\r\n")},
			wantContent: "<p>a</p>\r\n",
		},
		{
			name:        "an empty chain leaves the payload alone",
			chain:       nil,
			payload:     Payload{ContentType: "text/html", Content: []byte("<p>a</p>")},
			wantContent: "<p>a</p>",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := applyChain(t, tc.chain, tc.payload)

			if string(got.Content) != tc.wantContent {
				t.Errorf("Content = %q, want %q", got.Content, tc.wantContent)
			}
			if got.Extension != tc.wantExtension {
				t.Errorf("Extension = %q, want %q", got.Extension, tc.wantExtension)
			}
		})
	}
}

func TestApplyStopsOnTransformError(t *testing.T) {
	boom := errors.New("boom")
	reached := false

	chain := Chain{
		{Name: "failing", Accepts: []Kind{KindHTML}, Produces: KindSame,
			Transform: func(Payload) (Payload, error) { return Payload{}, boom }},
		{Name: "later", Accepts: []Kind{KindHTML}, Produces: KindSame,
			Transform: func(p Payload) (Payload, error) { reached = true; return p, nil }},
	}

	_, err := chain.Apply(Payload{ContentType: "text/html"})
	if !errors.Is(err, ErrTransformFailed) || !errors.Is(err, boom) {
		t.Fatalf("err = %v, want ErrTransformFailed wrapping boom", err)
	}
	if !strings.Contains(err.Error(), `"failing" at position 1`) {
		t.Errorf("error %q does not locate the failing mapper", err)
	}
	if reached {
		t.Error("chain continued past a failing mapper")
	}
}
