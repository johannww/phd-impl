package handlers

import (
	"crypto/ed25519"
	"encoding/json"

	"github.com/gin-gonic/gin"
	cc_auction "github.com/johannww/phd-impl/chaincodes/carbon/auction"
	"github.com/johannww/phd-impl/tee_auction/auction"
)

func Auction(c *gin.Context, privateKey ed25519.PrivateKey) {
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

	auctionResult, err := auction.RunTEEAuction(&serializedAD, privateKey)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to run auction: " + err.Error()})
		return
	}
	c.JSON(200, auctionResult)
}
