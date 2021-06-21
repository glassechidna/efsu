package main

import (
	"bytes"
	"context"
	_ "embed"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"io"
	"os"
	"strings"
	"time"
)

//go:embed bucket.yml
var bucketCfn string

//go:embed lambda.yml
var lambdaCfn string

//go:embed lambda.zip
var lambdaZip []byte

func doSetup(ctx context.Context, cfn cloudformationiface.CloudFormationAPI, s3api s3iface.S3API, api lambdaiface.LambdaAPI) error {
	cfg := ConfigFromContext(ctx)

	bucketsOutput, err := ensureStack(ctx, cfn, "efsu-bucket", bucketCfn, map[string]string{
		"BucketName": cfg.BucketName,
	})
	if err != nil {
		return err
	}

	bucket := bucketsOutput["Bucket"]
	key := "efsu.zip"

	put, err := s3api.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   bytes.NewReader(lambdaZip),
	})
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = ensureStack(ctx, cfn, "efsu-lambda", lambdaCfn, map[string]string{
		"FunctionName":      cfg.FunctionName,
		"SubnetId":          cfg.SubnetId,
		"SecurityGroupId":   cfg.SecurityGroupId,
		"EfsAccessPointArn": cfg.AccessPointArn,
		"Bucket":            bucket,
		"Key":               key,
		"VersionId":         *put.VersionId,
	})
	if err != nil {
		return err
	}

	return nil
}

func ensureStack(ctx context.Context, cfn cloudformationiface.CloudFormationAPI, name, template string, params map[string]string) (map[string]string, error) {
	parameters := []*cloudformation.Parameter{}
	for k, v := range params {
		parameters = append(parameters, &cloudformation.Parameter{
			ParameterKey:   aws.String(k),
			ParameterValue: aws.String(v),
		})
	}

	capabilities := aws.StringSlice([]string{
		cloudformation.CapabilityCapabilityIam,
		cloudformation.CapabilityCapabilityAutoExpand,
	})

	tags := []*cloudformation.Tag{
		{Key: aws.String("efsu:version"), Value: aws.String(version)},
	}

	exists, err := stackExists(ctx, cfn, name)
	if err != nil {
		return nil, err
	}

	eventsSince := time.Now()

	if exists {
		_, err = cfn.UpdateStackWithContext(ctx, &cloudformation.UpdateStackInput{
			StackName:    &name,
			TemplateBody: &template,
			Parameters:   parameters,
			Capabilities: capabilities,
			Tags:         tags,
		})
		if err != nil {
			return nil, errors.WithStack(err)
		}
	} else {
		_, err = cfn.CreateStackWithContext(ctx, &cloudformation.CreateStackInput{
			StackName:    &name,
			TemplateBody: &template,
			Parameters:   parameters,
			Capabilities: capabilities,
			Tags:         tags,
		})
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	for {
		time.Sleep(time.Second)
		describe, err := cfn.DescribeStacksWithContext(ctx, &cloudformation.DescribeStacksInput{StackName: &name})
		if err != nil {
			return nil, errors.WithStack(err)
		}

		stack := describe.Stacks[0]
		w := os.Stdout

		eventsSince = printEventsSince(ctx, cfn, name, eventsSince, w)

		if terminalStatus(*stack.StackStatus) {
			outputs := map[string]string{}
			for _, output := range stack.Outputs {
				outputs[*output.OutputKey] = *output.OutputValue
			}

			if !successStatus(*stack.StackStatus) {
				return nil, errors.Errorf("unsuccessful stack operation: %s", *stack.StackStatusReason)
			}

			return outputs, nil
		}
	}
}

func printEventsSince(ctx context.Context, cfn cloudformationiface.CloudFormationAPI, name string, eventsSince time.Time, w io.Writer) time.Time {
	newEvents := []*cloudformation.StackEvent{}

	input := &cloudformation.DescribeStackEventsInput{StackName: &name}
	err := cfn.DescribeStackEventsPagesWithContext(ctx, input, func(page *cloudformation.DescribeStackEventsOutput, lastPage bool) bool {
		for _, event := range page.StackEvents {
			if (*event.Timestamp).After(eventsSince) {
				newEvents = append(newEvents, event)
			}
		}

		lastTimestamp := page.StackEvents[len(page.StackEvents)-1].Timestamp
		return !lastPage || lastTimestamp.Before(eventsSince)
	})
	if err != nil {
		panic(err)
	}

	table := tablewriter.NewWriter(w)
	table.SetTrimWhiteSpaceAtEOL(true)
	table.SetAutoWrapText(false)
	table.SetColumnAlignment([]int{
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_LEFT,
	})
	table.SetColMinWidth(1, 26)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding(" ") // pad with tabs
	table.SetNoWhiteSpace(true)

	for idx := len(newEvents) - 1; idx >= 0; idx-- {
		event := newEvents[idx]
		ts := *event.Timestamp
		eventsSince = ts

		date := ts.Format("[15:04:05]")
		status := *event.ResourceStatus
		if event.ResourceStatusReason != nil {
			status += " " + *event.ResourceStatusReason
		}
		table.Append([]string{date, *event.LogicalResourceId, "-", status})
	}

	table.Render()
	return eventsSince
}

func stackExists(ctx context.Context, cfn cloudformationiface.CloudFormationAPI, name string) (bool, error) {
	_, err := cfn.DescribeStacksWithContext(ctx, &cloudformation.DescribeStacksInput{StackName: &name})
	if err != nil {
		if err, ok := err.(awserr.Error); ok {
			if err.Code() == "ValidationError" && strings.Contains(err.Message(), "does not exist") {
				return false, nil
			}
		}

		return false, errors.WithStack(err)
	}

	return true, nil
}

func terminalStatus(status string) bool {
	switch status {
	case
		"CREATE_COMPLETE",
		"DELETE_COMPLETE",
		"CREATE_FAILED",
		"DELETE_FAILED",
		"ROLLBACK_COMPLETE",
		"ROLLBACK_FAILED",
		"UPDATE_COMPLETE",
		"UPDATE_FAILED",
		"UPDATE_ROLLBACK_COMPLETE",
		"UPDATE_ROLLBACK_FAILED":
		return true
	default:
		return false
	}
}

func successStatus(status string) bool {
	return status == "CREATE_COMPLETE" || status == "UPDATE_COMPLETE"
}
