package main

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var cfgFile string

type Config struct {
	FunctionName    string
	AccessPointArn  string
	SubnetId        string
	SecurityGroupId string
	BucketName      string
	YOLO            bool
	session         *session.Session
}

type configKey struct{}

func ConfigFromContext(ctx context.Context) Config {
	return ctx.Value(configKey{}).(Config)
}

func withConfig(cmd *cobra.Command) (context.Context, Config) {
	profile := viper.GetString("profile")
	region := viper.GetString("region")

	var logLevel aws.LogLevelType
	if _, ok := os.LookupEnv("AWS_VERBOSE_LOG"); ok {
		logLevel = aws.LogDebugWithHTTPBody
	}

	config := *aws.NewConfig().
		WithRegion(region).
		WithLogLevel(logLevel)

	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState:       session.SharedConfigEnable,
		AssumeRoleTokenProvider: stscreds.StdinTokenProvider,
		Profile:                 profile,
		Config:                  config,
	})
	if err != nil {
		panic(err)
	}

	cfg := Config{
		FunctionName:    viper.GetString("function-name"),
		AccessPointArn:  viper.GetString("access-point-arn"),
		SubnetId:        viper.GetString("subnet-id"),
		SecurityGroupId: viper.GetString("security-group-id"),
		BucketName:      viper.GetString("bucket-name"),
		YOLO:            viper.GetBool("yolo"),
		session:         sess,
	}

	return context.WithValue(cmd.Context(), configKey{}, cfg), cfg
}
