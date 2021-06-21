package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/glassechidna/efsu"
	"github.com/olekukonko/tablewriter"
	"io"
	"math"
)

func doLs(ctx context.Context, api *lambda.Lambda, path string, recursive bool, out io.Writer) {
	input := &efsu.Input{
		Command: efsu.CommandList,
		List: &efsu.ListInput{
			Path:      path,
			Recursive: recursive,
		},
	}

	for {
		output, err := invoke(ctx, api, input)
		if err != nil {
			panic(err)
		}

		renderItemList(output, out)

		next := output.List.Next
		input.List.Next = next
		if next == "" {
			break
		}
	}

}

func renderItemList(output *efsu.Output, w io.Writer) {
	table := tablewriter.NewWriter(w)
	table.SetTrimWhiteSpaceAtEOL(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetColumnAlignment([]int{
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_LEFT,
	})
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t") // pad with tabs
	table.SetNoWhiteSpace(true)

	for _, item := range output.List.Items {
		date := item.ModTime.Format("02 Jan 2006")
		table.Append([]string{item.Mode.String(), friendlySize(item.Size, item.Mode.IsDir()), date, item.Path})
	}

	table.Render()
}

func friendlySize(bytes int64, isDir bool) string {
	if isDir {
		return "-"
	}

	flt := float64(bytes)
	if flt > 1e9 {
		return fmt.Sprintf("%0.1fG", flt/math.Pow(2, 30))
	} else if flt > 1e6 {
		return fmt.Sprintf("%0.1fM", flt/math.Pow(2, 20))
	} else if flt > 1e3 {
		return fmt.Sprintf("%0.1fK", flt/math.Pow(2, 10))
	} else {
		return fmt.Sprintf("%d", bytes)
	}
}
