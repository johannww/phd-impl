package handlers

import (
	"crypto/ed25519"
	"encoding/json"

	"github.com/gin-gonic/gin"
	cc_auction "github.com/johannww/phd-impl/chaincodes/carbon/auction"
	"github.com/johannww/phd-impl/tee_auction/go/auction"
)

func Auction(c *gin.Context, privateKey ed25519.PrivateKey, certDer []byte) {
	// Get the user name and value from the request body
	dataBytes, err := c.GetRawData()
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	var serializedAD cc_auction.SerializedAuctionData
	err = json.Unmarshal(dataBytes, &serializedAD)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid auction data"})
		return
	}

	auctionResultPub, auctionResultPvt, err := auction.RunTEEAuction(&serializedAD, privateKey)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to run auction: " + err.Error()})
		return
	}
	auctionResultPub.TEECertDer = certDer
	auctionResultPvt.TEECertDer = certDer
	c.JSON(200, gin.H{"public": auctionResultPub, "private": auctionResultPvt})
}
