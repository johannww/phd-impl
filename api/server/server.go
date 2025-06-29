package server

import (
	"github.com/gin-gonic/gin"
	"github.com/johannww/phd-impl/api/api"
	"github.com/johannww/phd-impl/api/handlers"
)

type CarbonAPIServer struct {
	TeeHandler   *handlers.TeeHandler
	ChainHandler *handlers.ChainHandler
}

var _ api.ServerInterface = (*CarbonAPIServer)(nil)

func NewCarbonAPIServer() *CarbonAPIServer {
	return &CarbonAPIServer{
		TeeHandler:   &handlers.TeeHandler{},
		ChainHandler: &handlers.ChainHandler{},
	}
}

// GetTeeReport implements api.ServerInterface.
func (s *CarbonAPIServer) GetTeeReport(c *gin.Context) {
	s.TeeHandler.GetTeeReport(c)
}

// PostChainRegisterProperty implements api.ServerInterface.
func (s *CarbonAPIServer) PostChainRegisterProperty(c *gin.Context) {
	s.ChainHandler.PostChainRegisterProperty(c)
}

// PostTeeSetIp implements api.ServerInterface.
func (s *CarbonAPIServer) PostTeeSetIp(c *gin.Context, params api.PostTeeSetIpParams) {
	s.TeeHandler.PostTeeSetIp(c, params)
}
