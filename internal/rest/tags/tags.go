package tags

import (
	"ClipsArchiver/internal/db"
	"ClipsArchiver/internal/rest"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetAll(c *gin.Context) {
	tags, err := db.GetAllTags()
	if err != nil {
		c.String(http.StatusInternalServerError, rest.Error500String)
		return
	}
	c.IndentedJSON(http.StatusOK, tags)
}
