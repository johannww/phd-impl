package main

import (
	"github.com/johannww/phd-impl/chaincodes/carbon/application"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// This is an application using the fabric-gateway-go to interact with
// the carbon chaincode

func main() {
	pflag.String("mspId", "", "MSP ID")
	pflag.String("mspPath", "", "MSP path")
	pflag.Bool("idemix", false, "use idemix credentials")
	pflag.Parse()

	viper.BindPFlags(pflag.CommandLine)

	mspID := viper.GetString("mspId")
	mspPath := viper.GetString("mspPath")
	idemix := viper.GetBool("idemix")

	if mspPath == "" {
		panic("mspPath cannot be nil")
	}

	application.Run(idemix, mspPath, mspID)
}
