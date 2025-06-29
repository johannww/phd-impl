package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/johannww/phd-impl/api/api"
)

type TeeHandler struct {
	TeeIP string `json:"tee_ip,omitempty"`
}

func (t *TeeHandler) PostTeeSetIp(c *gin.Context, params api.PostTeeSetIpParams) {
	t.TeeIP = params.Ip
	c.JSON(200, gin.H{
		"message": "TEE IP set successfully",
		"tee_ip":  t.TeeIP,
	})
}

func (t *TeeHandler) GetTeeReport(c *gin.Context) {
	log.Default().Println("TEE Report requested")

	resp, err := http.Get(fmt.Sprintf("http://%s:8080/report", t.TeeIP))
	if err != nil {
		log.Default().Printf("Failed to get TEE report: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get TEE report"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Default().Printf("TEE report request failed with status: %s", resp.Status)
		c.JSON(resp.StatusCode, gin.H{"error": "Failed to get TEE report"})
		return
	}

	report, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Default().Printf("Failed to read TEE report response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read TEE report response"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"report": report})

}
