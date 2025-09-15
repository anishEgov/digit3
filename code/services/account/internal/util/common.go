package util

import (
	"account/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GetResponseInfo(c *gin.Context, status string) models.ResponseInfo {
	return models.ResponseInfo{
		APIId:  c.FullPath(),
		Ver:    "1.0",
		Ts:     time.Now().Unix(),
		Status: status,
		MsgId:  c.GetHeader("X-Client-Id"),
	}
}

func GenerateUUID() string {
	return uuid.New().String()
}
