package clips

import (
	"ClipsArchiver/internal/db"
	"ClipsArchiver/internal/rest"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func Update(c *gin.Context) {
	var clip db.Clip
	if err := c.BindJSON(&clip); err != nil {
		c.String(http.StatusBadRequest, "Invalid body for Clip")
	}
	clipId, conversionErr := strconv.Atoi(c.Param("clipId"))
	if conversionErr != nil {
		c.String(http.StatusBadRequest, "invalid clip id provided: %s", c.Param("clipId"))
		return
	}
	existingClip, err := db.GetClipById(clipId)
	if err != nil {
		c.String(http.StatusInternalServerError, rest.Error500String)
		return
	}
	err = db.UpdateClipTags(existingClip, clip)
	if err != nil {
		c.String(http.StatusInternalServerError, rest.Error500String)
	}
	err = db.UpdateClip(clip)
	if err != nil {
		c.String(http.StatusInternalServerError, rest.Error500String)
	}
}

func Delete(c *gin.Context) {
	clipId, err := strconv.Atoi(c.Param("clipId"))
	if err != nil {
		c.String(http.StatusBadRequest, "invalid clip id provided: %s", c.Param("clipId"))
		return
	}
	err = db.DeleteClipById(clipId)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid clip id provided: %s", c.Param("clipId"))
		return
	}
	c.String(http.StatusNoContent, "Deleted clip")
}

func Get(c *gin.Context) {
	clipId, err := strconv.Atoi(c.Param("clipId"))
	if err != nil {
		c.String(http.StatusBadRequest, "invalid clip id provided: %s", c.Param("clipId"))
		return
	}
	clip, err := db.GetClipById(clipId)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid clip id provided: %s", c.Param("clipId"))
		return
	}
	c.IndentedJSON(http.StatusOK, clip)
}

func GetForDate(c *gin.Context) {
	date := c.Param("date")
	values := strings.Split(date, "-")
	if len(values) != 3 {
		c.String(http.StatusBadRequest, "Invalid date format: Should be YYYY-MM-DD.")
		return
	}
	year, err := strconv.Atoi(values[0])
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid date format: Should be YYYY-MM-DD.")
		return
	}

	month, err := strconv.Atoi(values[1])
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid date format: Should be YYYY-MM-DD.")
		return
	}

	day, err := strconv.Atoi(values[2])
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid date format: Should be YYYY-MM-DD.")
		return
	}

	clips, err := db.GetClipsForDate(time.Date(year, time.Month(month), day, 4, 0, 0, 0, time.UTC))

	if err != nil {
		println(err)
		c.String(http.StatusInternalServerError, rest.Error500String)
	}

	c.IndentedJSON(http.StatusOK, clips)
}

func GetByFilename(c *gin.Context) {
	clip, err := db.GetClipByFilename(c.Param("filename"))
	if err != nil {
		println(err.Error())
		c.String(http.StatusNotFound, "No clip found for filename")
		return
	}
	c.IndentedJSON(http.StatusOK, clip)
}
