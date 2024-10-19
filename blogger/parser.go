package blogger

import (
	"fmt"
	"os"
	"strings"

	"github.com/adrg/frontmatter"
)

func Parse(fname string) (*Post, error) {
	fmt.Printf("[*] parsing %s\n", fname)

	data, err := os.ReadFile(fname)
	if err != nil {
		return nil, fmt.Errorf("read data: %w", err)
	}

	post := &Post{}
	rest, err := frontmatter.Parse(strings.NewReader(string(data)), post)
	if err != nil {
		return nil, fmt.Errorf("metadata read: %w", err)
	}

	post.Content = string(rest)

	post.File = fname

	return post, nil
}
