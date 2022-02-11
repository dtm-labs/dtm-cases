package utils

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/dtm-labs/dtmcli"
	"github.com/dtm-labs/dtmcli/logger"
	"github.com/gin-gonic/gin"
)

func getStatusFromResult(result interface{}) int {
	err, _ := result.(error)
	if errors.Is(err, dtmcli.ErrFailure) {
		return http.StatusConflict
	} else if errors.Is(err, dtmcli.ErrOngoing) {
		return http.StatusTooEarly
	} else if err != nil {
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

func WrapHandler(fn func(*gin.Context) interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		began := time.Now()
		ret := fn(c)
		status := getStatusFromResult(ret)

		b, _ := json.Marshal(ret)
		if status == http.StatusOK || status == http.StatusTooEarly {
			logger.Infof("%2dms %d %s %s %s", time.Since(began).Milliseconds(), status, c.Request.Method, c.Request.RequestURI, string(b))
		} else {
			logger.Errorf("%2dms %d %s %s %s", time.Since(began).Milliseconds(), status, c.Request.Method, c.Request.RequestURI, string(b))
		}
		c.JSON(status, ret)
	}
}
