package main

import (
	"bytes"
	"context"
	"github.com/glassechidna/efsu"
	"github.com/pkg/errors"
	"io"
	"os"
)

func handleDownload(ctx context.Context, input *efsu.Input) (*efsu.Output, error) {
	path := input.Download.Path
	info, err := os.Stat(path)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	fRange := input.Download.Range
	if fRange.Size == 0 {
		fRange.Size = info.Size()
	}

	var limit int64 = 4e6
	if fRange.Size > limit {
		fRange.Size = limit
	}

	_, err = f.Seek(fRange.Offset, io.SeekStart)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if fRange.Offset+fRange.Size > info.Size() {
		fRange.Size = info.Size() - fRange.Offset
	}

	buf := &bytes.Buffer{}
	copied, err := io.CopyN(buf, f, fRange.Size)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if copied != fRange.Size {
		return nil, errors.New("too little data read")
	}

	return &efsu.Output{
		Version: version,
		Download: &efsu.DownloadOutput{
			FileSize: info.Size(),
			Mode:     info.Mode(),
			Content:  buf.Bytes(),
		},
	}, nil
}
