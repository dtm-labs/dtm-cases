package demo

import (
	"github.com/dtm-labs/dtm-cases/utils"
	"github.com/gin-gonic/gin"
)

func addStringConsistency(app *gin.Engine) {
	app.GET(BusiAPI+"/scDemo", utils.WrapHandler(func(c *gin.Context) interface{} {
		return "nil"
	}))
}
