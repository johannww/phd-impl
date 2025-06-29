package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/johannww/phd-impl/api/api"
	"github.com/johannww/phd-impl/api/server"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	pflag.String("mspId", "", "MSP ID")
	pflag.Parse()

	viper.BindPFlags(pflag.CommandLine)

	router := gin.Default()
	api.RegisterHandlers(router, server.NewCarbonAPIServer())

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		panic(fmt.Sprintf("Failed to start server: %v", err))
	}
}
