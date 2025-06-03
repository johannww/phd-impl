package main

import (
	"fmt"
	"os"

	"github.com/johannww/phd-impl/tee_auction/report"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	pflag.String("reportJsonPath", "", "Path to the attestation report JSON file")
	pflag.Parse()

	viper.BindPFlags(pflag.CommandLine)

	reportJsonPath := viper.GetString("reportJsonPath")

	reportJsonBytes, err := os.ReadFile(reportJsonPath)
	if err != nil {
		panic(fmt.Sprintf("Failed to read report JSON file: %v", err))
	}

	verify, err := report.VerifyReportSignature(reportJsonBytes)

	if err != nil {
		panic(fmt.Sprintf("Failed to verify report signature: %v", err))
	}

	if verify {
		fmt.Println("Report signature verification succeeded.")
	} else {
		fmt.Println("Report signature verification failed.")
	}
}
