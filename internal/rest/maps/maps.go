package maps

import (
	"ClipsArchiver/internal/db"
	"ClipsArchiver/internal/rest"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetAll(c *gin.Context) {
	gameMaps, err := db.GetAllMaps()
	if err != nil {
		c.String(http.StatusInternalServerError, rest.ErrorDefault)
		return
	}
	c.IndentedJSON(http.StatusOK, gameMaps)
}
