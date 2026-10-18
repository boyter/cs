package search

import (
	"reflect"
	"strings"
	"testing"
)

// parseQuery parses query and fails the test on a parser notice.
func parseQuery(t *testing.T, query string) Node {
	t.Helper()
	ast, notices := NewParser(NewLexer(strings.NewReader(query))).ParseQuery()
	if len(notices) > 0 {
		t.Fatalf("%q: unexpected notices %v", query, notices)
	}
	return ast
}

// The cases below are queries people wrote against a real estate that came
// back empty or wrong (searchcode bugs/513).

// A colon filter takes a comma list, as the operator form always did. The
// lexer ends "ext:md" at the comma and the parser used to stop there, so the
// rest of the list was dropped silently.
func TestAColonFilterTakesACommaList(t *testing.T) {
	ast := parseQuery(t, "ext:md,yaml,json")
	f, ok := ast.(*FilterNode)
	if !ok {
		t.Fatalf("got %T (%v), want a FilterNode", ast, ast)
	}
	if want := []interface{}{"md", "yaml", "json"}; !reflect.DeepEqual(f.Value, want) {
		t.Errorf("value = %#v, want %#v", f.Value, want)
	}

	if f := parseQuery(t, "ext:md").(*FilterNode); f.Value != "md" {
		t.Errorf("a single value = %#v, want the string md", f.Value)
	}
}

// ...and so did every term after it: "ext:ts,sql zzznope" searched for ext:ts
// alone and returned files that do not contain zzznope.
func TestTermsAfterAColonListAreKept(t *testing.T) {
	ast := parseQuery(t, "ext:ts,sql zzznope")
	and, ok := ast.(*AndNode)
	if !ok {
		t.Fatalf("got %T (%v), want the list AND the keyword", ast, ast)
	}
	if k, ok := and.Right.(*KeywordNode); !ok || k.Value != "zzznope" {
		t.Errorf("right = %v, want KEYWORD(zzznope)", and.Right)
	}

	matched, _ := EvaluateFile(ast, []byte("export const x = 1"), "roles.ts", "apps/auth/roles.ts", false)
	if matched {
		t.Error("a .ts file without zzznope matched")
	}
}

// A list is "any of" in per-file evaluation, for ext, file and path, and "none
// of" with !=. It used to match every file, so "ext=sql,py" matched a .ts file.
func TestPerFileEvaluationHonoursFilterLists(t *testing.T) {
	cases := []struct {
		query    string
		location string
		want     bool
	}{
		{"ext=sql,py", "apps/auth/roles.ts", false},
		{"ext=sql,py", "apps/auth/schema.sql", true},
		{"ext:ts,tsx", "apps/web/index.tsx", true},
		{"ext:ts,tsx", "apps/web/index.css", false},
		{"ext!=ts,tsx", "apps/web/index.tsx", false},
		{"ext!=ts,tsx", "apps/auth/schema.sql", true},
		{"file=README,CLAUDE", "CLAUDE.md", true},
		{"file=README,CLAUDE", "main.go", false},
		{"path=apps/auth,libs", "libs/pyrocell_wrapper/src/client.py", true},
		{"path=apps/auth,libs", "apps/web/index.tsx", false},
		{"path!=tenant_api,geoserver", "apps/tenant_api/main.py", false},
		{"path!=tenant_api,geoserver", "apps/auth/roles.ts", true},
	}
	for _, c := range cases {
		ast := parseQuery(t, c.query)
		filename := c.location[strings.LastIndex(c.location, "/")+1:]
		if got, _ := EvaluateFile(ast, []byte("x"), filename, c.location, false); got != c.want {
			t.Errorf("%q on %s = %v, want %v", c.query, c.location, got, c.want)
		}
	}
}

// The document engine's path filter refused a list outright.
func TestTheEnginePathFilterHonoursAList(t *testing.T) {
	val := []interface{}{"apps/auth", "libs"}
	if !handlePathFilter("=", val, &Document{Path: "libs/x/client.py"}) {
		t.Error("path=apps/auth,libs did not match libs/x/client.py")
	}
	if handlePathFilter("=", val, &Document{Path: "apps/web/index.tsx"}) {
		t.Error("path=apps/auth,libs matched apps/web/index.tsx")
	}
	if handlePathFilter("!=", val, &Document{Path: "libs/x/client.py"}) {
		t.Error("path!=apps/auth,libs matched libs/x/client.py")
	}
}

// ^ and $ mean the line. Search runs over a whole file, so a file-anchored ^
// matched only a file that began with the pattern.
func TestARegexIsAnchoredToTheLine(t *testing.T) {
	content := []byte("stages:\n  - build:\n      steps:\n        - Name: compile\n")
	for _, c := range []struct {
		query string
		want  bool
	}{
		{`/^\s{8}- Name: /`, true},
		{`/compile$/`, true},
		{`/\A\s{8}- Name: /`, false},
		{`/\Astages:/`, true},
	} {
		if got, _ := EvaluateFile(parseQuery(t, c.query), content, "meta-pipeline.yaml", "pipelines/meta-pipeline.yaml", false); got != c.want {
			t.Errorf("%s = %v, want %v", c.query, got, c.want)
		}
	}
}

// A regex follows the search's case setting like every other term, and a
// pattern's own flag still wins.
func TestARegexFollowsTheCaseSetting(t *testing.T) {
	content := []byte("API_KEY=\"abcdefghijklmnop1234\"\n")
	for _, c := range []struct {
		query         string
		caseSensitive bool
		want          bool
	}{
		{`/api_key/`, false, true},
		{`/api_key/`, true, false},
		{`/API_KEY/`, true, true},
		{`/(?-i)api_key/`, false, false},
	} {
		if got, _ := EvaluateFile(parseQuery(t, c.query), content, ".env", ".env", c.caseSensitive); got != c.want {
			t.Errorf("%s caseSensitive=%v = %v, want %v", c.query, c.caseSensitive, got, c.want)
		}
	}

	engine := NewSearchEngine([]*Document{{Filename: ".env", Path: ".env", Content: content}})
	if res, _ := engine.Search(`/api_key/`, false); len(res.Documents) != 1 {
		t.Errorf("the document engine found %d documents for /api_key/ case-insensitively, want 1", len(res.Documents))
	}
}

// An invalid pattern's error quotes what the user typed, not the flag group
// the compiler adds.
func TestAnInvalidRegexErrorQuotesThePatternAsTyped(t *testing.T) {
	n := &RegexNode{Pattern: "(abc"}
	err := n.Compile()
	if err == nil {
		t.Fatal("(abc compiled")
	}
	if strings.Contains(err.Error(), "(?m") {
		t.Errorf("the error shows a flag the user never typed: %v", err)
	}
	if n.RegexpFor(false) != nil {
		t.Error("an invalid pattern returned a case-folded regexp")
	}
}
