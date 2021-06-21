package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/glassechidna/efsu"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"io"
	"os"
	"path/filepath"
)

func doCp(ctx context.Context, api *lambda.Lambda, src string, dst string) error {
	input := &efsu.Input{
		Command: efsu.CommandDownload,
		Download: &efsu.DownloadInput{
			Path:  src,
			Range: efsu.ContentRange{},
		},
	}

	output, err := invoke(ctx, api, input)
	if err != nil {
		return err
	}

	dl := output.Download

	dst, _ = homedir.Expand(dst)
	dst, _ = filepath.Abs(dst)
	dstStat, _ := os.Stat(dst)
	isDir := dstStat != nil && dstStat.IsDir()
	if isDir {
		dst = filepath.Join(dst, filepath.Base(src))
	}

	f, err := os.Create(dst)
	if err != nil {
		return errors.WithStack(err)
	}
	defer f.Close()

	err = f.Truncate(dl.FileSize)
	if err != nil {
		return errors.WithStack(err)
	}

	var written int64 = 0
	for {
		n, err := io.Copy(f, bytes.NewReader(dl.Content))
		if err != nil {
			return errors.WithStack(err)
		}

		written += n
		if written == dl.FileSize {
			break
		}

		fmt.Printf("now asking for [%d,) (got %d of %d)\n", written, written, dl.FileSize)
		input.Download.Range.Offset = written

		output, err = invoke(ctx, api, input)
		if err != nil {
			panic(err)
		}
		dl = output.Download
	}

	return nil
}
