package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/glassechidna/efsu"
)

var version = "dev"

func main() {
	lambda.Start(handleCommand)
}

func handleCommand(ctx context.Context, input *efsu.Input) (*efsu.Output, error) {
	fmt.Printf("client=%s lambda=%s yolo=%t\n", input.Version, version, input.YOLO)
	if input.Version != version && !input.YOLO {
		return &efsu.Output{Version: version}, nil
	}

	switch input.Command {
	case efsu.CommandList:
		return handleList(ctx, input)
	case efsu.CommandDownload:
		return handleDownload(ctx, input)
	default:
		panic("unknown command")
	}
	return nil, nil
}

