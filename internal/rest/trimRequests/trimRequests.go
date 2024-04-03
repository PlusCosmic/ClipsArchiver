package trimRequests

import (
	"ClipsArchiver/internal/db"
	"ClipsArchiver/internal/rest"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetAll(c *gin.Context) {
	queueEntries, err := db.GetAllTrimRequests()
	if err != nil {
		c.String(http.StatusInternalServerError, rest.ErrorDefault)
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

	queueEntry, err := db.GetTrimRequestByClipId(clipId)

	if err != nil {
		println(err.Error())
		c.String(http.StatusBadRequest, "invalid clip id provided: %s", c.Param("clipId"))
		return
	}

	c.IndentedJSON(http.StatusOK, queueEntry)
}

func Create(c *gin.Context) {
	var queueEntry db.TrimRequest
	if err := c.BindJSON(&queueEntry); err != nil {
		c.String(http.StatusBadRequest, "invalid request body")
		return
	}

	err := db.CreateTrimRequest(queueEntry)
	if err != nil {
		c.String(http.StatusInternalServerError, rest.ErrorDefault)
		return
	}

	c.String(http.StatusCreated, "created")
}
