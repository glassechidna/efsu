package main

import (
	"context"
	"encoding/json"
	"github.com/glassechidna/efsu"
	"github.com/pkg/errors"
	"io/fs"
	"path/filepath"
	"sort"
)

func handleList(ctx context.Context, input *efsu.Input) (*efsu.Output, error) {
	items := []efsu.ListItem{}

	err := filepath.WalkDir(input.List.Path, func(path string, d fs.DirEntry, err error) error {
		info, err := d.Info()
		if err != nil {
			return nil
			//return errors.WithStack(err)
		}

		items = append(items, efsu.ListItem{
			Path:    path,
			Mode:    info.Mode(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		})

		if !input.List.Recursive && d.IsDir() && path != input.List.Path {
			return filepath.SkipDir
		}

		return nil
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	sort.SliceStable(items, func(i, j int) bool {
		return items[i].Path < items[j].Path
	})

	start := sort.Search(len(items), func(i int) bool {
		return items[i].Path >= input.List.Next
	})

	items = items[start:]

	next := ""
	raw, _ := json.Marshal(items)
	for len(raw) > 5e6 {
		count := len(items)
		newCount := int(0.9 * float64(count))
		next = items[newCount].Path
		items = items[:newCount]
		raw, _ = json.Marshal(items)
	}

	output := &efsu.Output{
		Version: version,
		List: &efsu.ListOutput{
			Items: items,
			Next:  next,
		},
	}

	return output, nil
}
