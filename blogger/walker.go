package blogger

import (
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
)

func TraverseDir(dir string) ([]Post, error) {
	fmt.Printf("[*] traversing %s\n", dir)

	posts := make([]Post, 0)

	if err := filepath.Walk(dir, func(ppath string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			if path.Ext(info.Name()) != ".md" {
				return nil
			}

			post, err := Parse(ppath)
			if err != nil {
				return fmt.Errorf("parse %v", err)
			}

			posts = append(posts, *post)

			return nil
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("walk: %w", err)
	}

	return posts, nil
}
