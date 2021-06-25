module github.com/glassechidna/efsu

go 1.16

require (
	github.com/aws/aws-lambda-go v1.24.0
	github.com/aws/aws-sdk-go v1.38.64
	github.com/klauspost/compress v1.13.1
	github.com/mitchellh/go-homedir v1.1.0
	github.com/olekukonko/tablewriter v0.0.5
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.7.0 // indirect
)

replace github.com/olekukonko/tablewriter => github.com/cockroachdb/tablewriter v0.0.5-0.20200105123400-bd15540e8847
