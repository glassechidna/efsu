package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/glassechidna/efsu"
	"github.com/klauspost/compress/zstd"
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

	_, err = f.Seek(input.Download.Offset, io.SeekStart)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var limit int = 4e6
	buf := &bytes.Buffer{}
	zw, err := zstd.NewWriter(buf)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var rawRead int64 = 0
	for buf.Len() < limit {
		copied, err := io.CopyN(zw, f, 1e6)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, errors.WithStack(err)
		}

		rawRead += copied
	}

	err = zw.Close()
	if err != nil {
	    return nil, errors.WithStack(err)
	}

	fmt.Printf("cp fileSize=%d buflen=%d read=%d\n", info.Size(), buf.Len(), rawRead)

	return &efsu.Output{
		Version: version,
		Download: &efsu.DownloadOutput{
			FileSize:   info.Size(),
			Mode:       info.Mode(),
			Content:    buf.Bytes(),
			NextOffset: rawRead + input.Download.Offset,
		},
	}, nil
}
