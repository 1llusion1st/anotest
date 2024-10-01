package anotest

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"oss.terrastruct.com/d2/d2compiler"
	"oss.terrastruct.com/d2/d2format"
	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2layouts/d2dagrelayout"
	"oss.terrastruct.com/d2/d2lib"
	"oss.terrastruct.com/d2/d2renderers/d2svg"
	"oss.terrastruct.com/d2/d2target"
	"oss.terrastruct.com/d2/lib/textmeasure"
)

type AnotateTestOpts struct {
	showDuration bool
}

type AnotateOpt func(o *AnotateTestOpts) error

func WithDuration() AnotateOpt {
	return func(o *AnotateTestOpts) error {
		o.showDuration = true

		return nil
	}
}

func NewAnotateTest(t *testing.T, fName string, opts ...AnotateOpt) (*AnotateTest, error) {
	if strings.HasPrefix(fName, "~/") {
		u, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("get home dir: %w", err)
		}

		fName = path.Join(u, fName[2:])
	}

	f, err := os.OpenFile(fName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}

	options := new(AnotateTestOpts)

	for _, opt := range opts {
		if err := opt(options); err != nil {
			return nil, err
		}
	}

	return &AnotateTest{
		t: t, f: f, options: options,
		durations: make(map[string]time.Duration),
	}, nil
}

type AnotatedTestFunc func(t *testing.T)

type AnotateTest struct {
	t       *testing.T
	f       *os.File
	options *AnotateTestOpts
	level   int

	path      []string
	durations map[string]time.Duration

	total   int
	fails   int
	history []string

	codeFile    string
	codeStart   int
	codeComment string
}

func (a *AnotateTest) Path() string {
	return strings.Join(a.path, " > ")
}

func (a *AnotateTest) Story(title string, ffunc AnotatedTestFunc) {
	fmt.Fprintf(a.f, "# %s\n\n\n", title)

	start := time.Now()

	success := a.t.Run(title, ffunc)

	end := time.Now()

	t := a.total
	f := a.fails
	s := t - f
	p := 100.0 / float64(t)

	sp := float64(s) * p
	fp := 100 - sp

	d := decimal.NewFromFloat

	summaryMsg := fmt.Sprintf(
		"# Summary(total: %d success: %d(%s %%) failed: %d(%s %%)) - %s",
		t, s, d(sp).Round(1).String(), f, d(fp).Round(1).String(), end.Sub(start).String(),
	)

	fmt.Fprintf(a.f, "%s\n\n", summaryMsg)

	for _, histItem := range a.history {
		fmt.Fprint(a.f, histItem)
	}

	if success {
		fmt.Printf("Congratulations!!!\n")
	}
}

func MakeStrPath(path string) string {
	return strings.ReplaceAll(path, " > ", "")
}

func MakeLink(title string, path string) string {
	return fmt.Sprintf(" [%s](#%s) ", title, path)
}

func (a *AnotateTest) ML(title string, path string) string {
	return MakeLink(title, path)
}

func (a *AnotateTest) MStrPath(path string) string {
	return MakeStrPath(path)
}

var key = 100

func (a *AnotateTest) PutD2Svg(d2DiadSource string) *AnotateTest {
	svg, err := a.GetD2Svg(d2DiadSource)
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile("/tmp/debug/1.test.svg", []byte(svg), 0644); err != nil {
		panic(err)
	}

	convertCmd := exec.Command("convert", "-density", "50", "/dev/stdin", "png:-")
	convertCmd.Stdin = strings.NewReader(svg)
	convertOutput, err := convertCmd.Output()
	if err != nil {
		fmt.Println("Error converting:", err)
		panic(err)
		return a
	}

	if err := os.WriteFile("/tmp/debug/2.converted.png", convertOutput, 0644); err != nil {
		panic(err)
	}

	// convertOutput, _ = os.ReadFile("/tmp/test2.png")

	escaped := url.PathEscape(string(convertOutput))

	if err := os.WriteFile("/tmp/debug/3.quoted.url", []byte(escaped), 0644); err != nil {
		panic(err)
	}

	imageURL := "![image](data:image/png;data," + escaped + ")"

	if err := os.WriteFile("/tmp/debug/4.result.image", []byte(imageURL), 0600); err != nil {
		panic(err)
	}

	fmt.Fprintf(a.f, "\n\n%s\n\n", imageURL)

	return a
}

func (a *AnotateTest) GetD2Svg(d2DiadSource string) (string, error) {
	logger := slog.Default()
	ctx := context.WithValue(context.Background(), &key, logger)

	ruler, err := textmeasure.NewRuler()
	if err != nil {
		return "", fmt.Errorf("new ruler: %w", err)
	}

	tmplAnalizer := `
	vars: {
	  d2-config: {
		layout-engine: elk
		sketch: true
	  }
	}

	direction: right

	%s
	`

	var (
		conf *d2target.Config
		g    *d2graph.Graph
	)

	g, conf, err = d2compiler.Compile("", strings.NewReader(fmt.Sprintf(tmplAnalizer, d2DiadSource)), nil)
	if err != nil {
		return "", fmt.Errorf("wrong template: %w", err)
	}

	cOpts := d2lib.CompileOptions{
		Ruler:  ruler,
		Layout: conf.LayoutEngine,
		LayoutResolver: func(engine string) (d2graph.LayoutGraph, error) {
			defaultLayout := func(ctx context.Context, g *d2graph.Graph) error {
				//nolint:wrapcheck
				return d2dagrelayout.Layout(ctx, g, nil)
			}

			return defaultLayout, nil
		},
	}

	sketch := true

	rOpts := d2svg.RenderOpts{
		Pad:    conf.Pad,
		Sketch: &sketch,
	}

	script := d2format.Format(g.AST)

	diagram, _, err := d2lib.Compile(ctx, script, &cOpts, &rOpts)
	if err != nil {
		return "", fmt.Errorf("d2lib.Compile: %w", err)
	}

	out, err := d2svg.Render(diagram, &rOpts)
	if err != nil {
		return "", fmt.Errorf("d2svg.Render: %w", err)
	}

	return string(out), nil

}

func (a *AnotateTest) StopCode(comment ...string) {
	text := a.codeComment
	if len(comment) > 0 {
		text = comment[0]
	}

	fmt.Fprintf(a.f, "%s\n\n```go\n", text)

	_, _, codeStop, ok := runtime.Caller(1)
	if !ok {
		return
	}

	code, err := os.ReadFile(a.codeFile)
	if err != nil {
		panic(err)
	}

	codeLines := strings.Split(string(code), "\n")

	codeLines = codeLines[a.codeStart : codeStop-1]

	maxLine := 0
	for _, line := range codeLines {
		if len(line) > maxLine {
			maxLine = len(line)
		}
	}

	minSpace := maxLine
	spaceLen := 0

	fmt.Printf("\nminSpace: %d\n", minSpace)

	skip := make(map[int]struct{}, len(codeLines))

	for i, line := range codeLines {
		l := len(line)

		lineNew := strings.TrimLeft(line, "\t ")
		lNew := len(lineNew)

		if lNew == 0 {
			skip[i] = struct{}{}

			continue
		}

		spaceLen = l - lNew
		fmt.Printf("space len: %d == (%d - %d = %d)\n", spaceLen, l, lNew, l-lNew)

		if spaceLen < minSpace {
			fmt.Printf("spaceL: %d minSpace: %d\n", spaceLen, minSpace)
			minSpace = spaceLen
		}

		fmt.Printf("line: '%s' lnew: '%s' l: %d lNew: %d space: %d minSpace: %d\n", line, lineNew, l, lNew, spaceLen, minSpace)
	}

	fmt.Printf("min space: %d\n", minSpace)

	for i := range codeLines {
		if _, ok := skip[i]; ok {
			codeLines[i] = ""

			continue
		}

		codeLines[i] = codeLines[i][minSpace:]
	}

	codeToPrint := strings.Join(codeLines, "\n")

	fmt.Fprintf(a.f, "%s\n```\n\n", codeToPrint)
}

func (a *AnotateTest) StartCode(comment ...string) {
	var ok bool

	if len(comment) > 0 {
		a.codeComment = comment[0]
	}

	_, a.codeFile, a.codeStart, ok = runtime.Caller(1)
	if !ok {
		return
	}

}

func (a *AnotateTest) Chapter(name, title string, ffunc AnotatedTestFunc) {
	a.total++
	a.level++
	a.path = append(a.path, name)

	defer func() {
		a.level -= 1
		a.path = a.path[:len(a.path)-1]
	}()

	fmt.Fprintf(a.f, "%s %s\n%s\n\n",
		strings.Repeat("#", a.level+1),
		name, title)

	start := time.Now()

	success := a.t.Run(name, ffunc)

	end := time.Now()

	icon := GSuccess

	if !success {
		icon = GFailed
		a.fails++
	}

	a.durations[a.Path()] = end.Sub(start)

	if a.options.showDuration {
		a.history = append(a.history, fmt.Sprintf("%s: %v - %s\n\n", a.Path(), icon, end.Sub(start).String()))
	} else {
		a.history = append(a.history, fmt.Sprintf("%s: %v\n\n", a.Path(), icon))
	}
}

func (a *AnotateTest) Comment(comment string) *AnotateTest {
	fmt.Fprintf(a.f, "%s\n\n", comment)

	return a
}

func (a *AnotateTest) Br() *AnotateTest {
	fmt.Fprintf(a.f, "\n<br/>\n")

	return a
}

// - suite.Chapter(name, title....)
// - suite.Subchapter(name, title ...)
// - suite.D2(name, description)
// - suite.Picture(name, tittle)
// - suite.Link()
// - suite.RestClient.... Anotate
