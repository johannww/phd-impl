package main

import (
	"encoding/json"
	"os"
	"time"

	utils_test "github.com/johannww/phd-impl/chaincodes/carbon/tests/utils"
)

func main() {
	filePath := os.Args[1]
	issueInterval := time.Duration(30) * time.Second

	testData := utils_test.GenData(
		10,
		3,
		5,
		"2023-01-01T00:00:00Z",
		"2023-01-01T00:30:00Z",
		issueInterval,
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
