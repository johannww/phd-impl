package main

import (
	"encoding/json"
	"os"

	utils_test "github.com/johannww/phd-impl/chaincodes/carbon/tests/utils"
)

func main() {
	filePath := os.Args[1]

	testData := utils_test.GenData(
		10,
		3,
		5,
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
