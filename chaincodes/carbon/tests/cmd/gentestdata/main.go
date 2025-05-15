package main

import (
	"encoding/json"
	"os"
	"time"

	utils_test "github.com/johannww/phd-impl/chaincodes/carbon/tests/utils"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	pflag.Int("owners", 10, "number of owners")
	pflag.Int("chunks", 3, "number of chunks")
	pflag.Int("companies", 5, "number of companies")
	pflag.String("start", "2023-01-01T00:00:00Z", "start timestamp")
	pflag.String("end", "2023-01-01T00:30:00Z", "end timestamp")
	pflag.Duration("interval", 30*time.Second, "issue interval")
	pflag.StringP("output", "o", "testdata.json", "output file path")
	pflag.Parse()

	viper.BindPFlags(pflag.CommandLine)

	filePath := viper.GetString("output")

	testData := utils_test.GenData(
		viper.GetInt("owners"),
		viper.GetInt("chunks"),
		viper.GetInt("companies"),
		viper.GetString("start"),
		viper.GetString("end"),
		viper.GetDuration("interval"),
	)

	bytes, err := json.MarshalIndent(testData, "", "  ")
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(filePath, bytes, os.ModePerm)
	if err != nil {
		panic(err)
	}
}
