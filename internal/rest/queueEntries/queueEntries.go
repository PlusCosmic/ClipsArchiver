package queueEntries

import (
	"ClipsArchiver/internal/db"
	"ClipsArchiver/internal/rest"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func GetAll(c *gin.Context) {
	queueEntries, err := db.GetClipsQueue()
	if err != nil {
		c.String(http.StatusInternalServerError, rest.Error500String)
		return
	}
	c.IndentedJSON(http.StatusOK, queueEntries)
}

func GetById(c *gin.Context) {
	clipId, conversionErr := strconv.Atoi(c.Param("clipId"))
	if conversionErr != nil {
		println(conversionErr.Error())
		c.String(http.StatusBadRequest, "invalid clip id provided: %s", c.Param("clipId"))
		return
	}

	queueEntry, err := db.GetQueueEntryByClipId(clipId)

	if err != nil {
		println(err.Error())
		c.String(http.StatusBadRequest, "invalid clip id provided: %s", c.Param("clipId"))
		return
	}

	c.IndentedJSON(http.StatusOK, queueEntry)
}
