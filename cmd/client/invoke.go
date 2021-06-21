package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/glassechidna/efsu"
	"github.com/pkg/errors"
	"os"
)

func invoke(ctx context.Context, api lambdaiface.LambdaAPI, input *efsu.Input) (*efsu.Output, error) {
	cfg := ConfigFromContext(ctx)

	input.Version = version
	input.YOLO = cfg.YOLO

	payload, err := json.Marshal(input)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	resp, err := api.InvokeWithContext(ctx, &lambda.InvokeInput{
		FunctionName: &cfg.FunctionName,
		Qualifier:    aws.String("live"),
		LogType:      aws.String(lambda.LogTypeTail),
		Payload:      payload,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if resp.FunctionError != nil {
		raw, _ := base64.StdEncoding.DecodeString(*resp.LogResult)
		fmt.Fprintln(os.Stderr, string(raw))
		return nil, errors.Errorf("lambda error: %s", *resp.FunctionError)
	}

	output := efsu.Output{}
	err = json.Unmarshal(resp.Payload, &output)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if output.Version != input.Version && !input.YOLO {
		return nil, errors.Errorf("version mismatch. local %s and lambda %s", input.Version, output.Version)
	}

	return &output, nil
}
