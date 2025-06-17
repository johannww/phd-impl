package api

import (
	"crypto/ed25519"
	"encoding/base64"
	"net/http"

	"github.com/Microsoft/confidential-sidecar-containers/pkg/attest"
	"github.com/gin-gonic/gin"
	"github.com/johannww/phd-impl/tee_auction/go/api/handlers"
	"github.com/johannww/phd-impl/tee_auction/go/report"
)

type AuctionServer struct {
	ReportBytes        []byte
	DeserializedReport *attest.SNPAttestationReport
}

var db = map[string]string{}

func (server *AuctionServer) SetupRouter(privateKey ed25519.PrivateKey) *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	r.GET("/report", func(c *gin.Context) {
		if server.DeserializedReport == nil {
			var err error
			server.DeserializedReport, err = report.DeserializedReport(server.ReportBytes)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to deserialize report"})
				return
			}
			server.DeserializedReport = server.DeserializedReport
		}

		c.JSON(http.StatusOK, server.DeserializedReport)
	})

	r.GET("/reportb64", func(c *gin.Context) {
		b64Str := base64.StdEncoding.EncodeToString(server.ReportBytes)
		c.JSON(http.StatusOK, b64Str)
	})

	r.POST("/auction", func(c *gin.Context) {
		handlers.Auction(c, privateKey)
	})

	// Get user value
	r.GET("/user/:name", func(c *gin.Context) {
		user := c.Params.ByName("name")
		value, ok := db[user]
		if ok {
			c.JSON(http.StatusOK, gin.H{"user": user, "value": value})
		} else {
			c.JSON(http.StatusOK, gin.H{"user": user, "status": "no value"})
		}
	})

	// Authorized group (uses gin.BasicAuth() middleware)
	// Same than:
	// authorized := r.Group("/")
	// authorized.Use(gin.BasicAuth(gin.Credentials{
	//	  "foo":  "bar",
	//	  "manu": "123",
	//}))
	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		"foo":  "bar", // user:foo password:bar
		"manu": "123", // user:manu password:123
	}))

	/* example curl for /admin with basicauth header
	   Zm9vOmJhcg== is base64("foo:bar")

		curl -X POST \
	  	http://localhost:8080/admin \
	  	-H 'authorization: Basic Zm9vOmJhcg==' \
	  	-H 'content-type: application/json' \
	  	-d '{"value":"bar"}'
	*/
	authorized.POST("admin", func(c *gin.Context) {
		user := c.MustGet(gin.AuthUserKey).(string)

		// Parse JSON
		var json struct {
			Value string `json:"value" binding:"required"`
		}

		_ = json
		_ = user

		// if c.Bind(&json) == nil {
		// 	db[user] = json.Value
		// 	c.JSON(http.StatusOK, gin.H{"status": "ok"})
		// }
	})

	return r
}
