package api

import "github.com/gin-gonic/gin"

type PingResponse struct {
	Message string `json:"message"`
}

func AddPingRoute(r *gin.Engine) {
	r.GET("/ping", func(c *gin.Context) {
		response := NewPingResponse("pong")
		c.JSON(200, response)
	})
}

func (pingResponse *PingResponse) GetMessage() string {
	return pingResponse.Message
}

func (pingResponse *PingResponse) SetMessage(message string) {
	pingResponse.Message = message
}

func NewPingResponse(message string) *PingResponse {
	return &PingResponse{
		Message: message,
	}
}
