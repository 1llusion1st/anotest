package blogger

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type Post struct {
	Author Author    `toml:"author"`
	Date   time.Time `toml:"date"`

	Title   string `toml:"title"`
	Tags    []Tag  `toml:"tags"`
	Summary string `toml:"summary"`
	Content string `toml:"content"`

	Metrics Metrics `toml:"metrics"`

	File string `toml:"-"`
}

func (p Post) String() string {
	return fmt.Sprintf("<Post %s [%v]: /%s>", p.Title, p.Tags, p.Slug())
}

func (p Post) GetContent() string {
	return base64.StdEncoding.EncodeToString([]byte(p.Content))
}

func (p Post) Slug() string {
	name := p.Date.Format("2024-10-02") + "-" + p.Title

	reg, _ := regexp.Compile("[^a-zA-Z0-9]+")

	processedString := reg.ReplaceAllString(name, " ")

	processedString = strings.TrimSpace(processedString)

	slug := strings.ReplaceAll(processedString, " ", "-")

	slug = strings.ToLower(slug)

	return slug
}

type Author struct {
	Name  string `toml:"name"`
	Email string `toml:"email"`
}

type Metrics struct {
	Words      int     `toml:"words"`
	Sentences  int     `toml:"sentences"`
	Cyclomatic float64 `toml:"cyclomatic"`
	Paragraphs int     `toml:"paragraphs"`
}

type Tag string
