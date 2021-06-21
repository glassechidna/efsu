package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"strings"
)

var version = "dev"
var commit, date string

func main() {
	cobra.OnInitialize(initConfig)

	root := &cobra.Command{
		Use:   "efsu",
		Short: "efsu is a simple way to interact with AWS EFS locally",
	}

	root.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "")
	root.PersistentFlags().String("function-name", "efsu", "")
	root.PersistentFlags().String("access-point-arn", "", "")
	root.PersistentFlags().String("subnet-id", "", "")
	root.PersistentFlags().String("security-group-id", "", "")
	root.PersistentFlags().String("profile", "", "")
	root.PersistentFlags().String("region", "", "")
	root.PersistentFlags().Bool("yolo", false, "")
	viper.BindPFlags(root.PersistentFlags())

	setup := &cobra.Command{
		Use:   "setup",
		Short: "Setup the utility Lambda function",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cfg := withConfig(cmd)
			sess := cfg.session
			err := doSetup(ctx, cloudformation.New(sess), s3.New(sess), lambda.New(sess))
			if err != nil {
				panic(err)
			}
		},
	}

	setup.PersistentFlags().String("bucket-name", "", "")
	viper.BindPFlags(setup.PersistentFlags())

	ls := &cobra.Command{
		Use:   "ls",
		Short: "List files on EFS",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cfg := withConfig(cmd)
			sess := cfg.session

			path := args[0]
			recursive, _ := cmd.Flags().GetBool("recursive")
			out := cmd.OutOrStdout()

			doLs(ctx, lambda.New(sess), path, recursive, out)
		},
	}

	ls.Flags().BoolP("recursive", "R", false, "")

	cp := &cobra.Command{
		Use:   "cp",
		Short: "Copy files from EFS to local system",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cfg := withConfig(cmd)
			sess := cfg.session

			src := args[0]
			dst := args[1]
			err := doCp(ctx, lambda.New(sess), src, dst)
			if err != nil {
			    panic(err)
			}
		},
	}

	versionCmd := &cobra.Command{
		Use: "version",
		Short: "Print CLI version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), `
Version: %s
Commit: %s
Build date: %s
`, version, commit, date)
		},
	}

	root.AddCommand(setup, ls, cp, versionCmd)

	ctx := context.Background()
	err := root.ExecuteContext(ctx)
	if err != nil {
		panic(err)
	}
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			panic(err)
		}

		viper.AddConfigPath(".")
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".efsu")
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix("EFSU")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	_ = viper.ReadInConfig()
}
