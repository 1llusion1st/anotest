package main

import (
	"bytes"
	"fmt"
	"os"
	"text/template"
	"time"

	"github.com/1llusion1st/anotest/blogger"
	"github.com/1llusion1st/anotest/blogger/static"
	"github.com/alecthomas/kong"
)

var commands = struct {
	Debug   bool
	Init    CmdInit    `cmd:""`
	Collect CmdCollect `cmd:""`
}{}

type Context struct {
	Debug bool
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := kong.Parse(&commands)

	err := ctx.Run(&Context{
		Debug: commands.Debug,
	})
	if err != nil {
		return fmt.Errorf("command run: %w", err)
	}

	fmt.Fprintf(os.Stdout, "success\n")

	return nil
}

type CmdInit struct{}

func (c *CmdInit) Run(ctx *Context) error {
	cwd, _ := os.Getwd()
	cwd = "."

	n := time.Now()

	year := n.Year()
	month := map[int]string{
		1:  "january",
		2:  "february",
		3:  "march",
		4:  "april",
		5:  "may",
		6:  "june",
		7:  "july",
		8:  "august",
		9:  "september",
		10: "october",
		11: "november",
		12: "december",
	}[int(n.Month())]

	err := os.MkdirAll(fmt.Sprintf("vault/%s/posts/%d/%s/\"\n", cwd, year, month), 0644)
	if err != nil {
		return fmt.Errorf("create storage: %w", err)
	}

	return nil
}

type CmdCollect struct {
	Path        string `name:"path"`
	Fact        string `name:"fact" default:"you can do everything!"`
	OutDir      string `name:"out" default:"public"`
	BrandPhrase string `name:"brand" default:"the best of the best"`
}

func (c *CmdCollect) Run(ctx *Context) error {
	if c.Path == "" {
		c.Path, _ = os.Getwd()
	}

	te := template.New("index")
	templ, err := te.Parse(string(static.Template))
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	fmt.Printf("[+] template loaded\n")

	posts, err := blogger.TraverseDir(c.Path)
	if err != nil {
		return fmt.Errorf("collect: %w", err)
	}

	fmt.Printf("[+] posts count: %d\n", len(posts))

	b := bytes.NewBuffer(nil)

	tagsDict := make(map[string]struct{})

	for _, post := range posts {
		fmt.Println(post)
		for _, tag := range post.Tags {
			tagsDict[string(tag)] = struct{}{}
		}
	}

	tags := make([]string, 0, len(tagsDict))

	for tag := range tagsDict {
		tags = append(tags, tag)
	}

	slugs := make([]string, 0, len(posts))
	for _, post := range posts {
		slugs = append(slugs, post.Slug())
	}

	postToShow := posts[len(posts)-1]

	err = templ.Execute(b, map[string]interface{}{
		"brand_phrase": c.BrandPhrase,

		"links_github":  "",
		"links_slate":   "",
		"links_discord": "",
		"links_twitter": "",
		"links_email":   "",

		"fact":  c.Fact,
		"tags":  tags,
		"slugs": slugs,
		"posts": posts,

		"CurrentSlug":    postToShow.Slug(),
		"CurrentContent": postToShow.GetContent(),
	})
	if err != nil {
		return fmt.Errorf("execute tamplate on index: %w", err)
	}

	err = os.WriteFile("index.html", b.Bytes(), 0644)

	return nil
}
