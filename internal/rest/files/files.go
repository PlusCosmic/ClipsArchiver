package files

import (
	"ClipsArchiver/internal/config"
	"ClipsArchiver/internal/db"
	"github.com/gin-gonic/gin"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func UploadClip(c *gin.Context) {
	ownerId, conversionErr := strconv.Atoi(c.Param("ownerId"))
	if conversionErr != nil {
		println(conversionErr.Error())
		c.String(http.StatusBadRequest, "invalid owner id provided: %s", c.Param("ownerId"))
		return
	}
	// Single file
	form, err := c.MultipartForm()
	files := form.File["file"]
	file := files[0]
	if err != nil {
		println(err.Error())
		c.String(http.StatusBadRequest, "get form err: %s", err.Error())
		return
	}

	clip, err := db.GetClipByFilename(file.Filename)

	if err == nil {
		println("clip exists")
		c.String(http.StatusBadRequest, "file already uploaded as clip with id: %d", clip.Id)
		return
	}

	filename := config.GetInputPath() + filepath.Base(file.Filename)
	if err := c.SaveUploadedFile(file, filename); err != nil {
		println(err.Error())
		c.String(http.StatusBadRequest, "upload file err: %s", err.Error())
		return
	}

	creationDateTime := form.Value["creationDateTime"][0]
	dateTimeParts := strings.Split(creationDateTime, "-")
	var dateTimePartsAsIntegers [6]int

	if len(dateTimeParts) != 6 {
		c.String(http.StatusBadRequest, "invalid creation date provided")
		return
	}

	for i := 0; i < 6; i++ {
		number, err := strconv.Atoi(dateTimeParts[i])
		if err != nil {
			println(err.Error())
			c.String(http.StatusBadRequest, "invalid creation date provided")
			return
		}
		dateTimePartsAsIntegers[i] = number
	}
	dateTime := time.Date(dateTimePartsAsIntegers[0], time.Month(dateTimePartsAsIntegers[1]), dateTimePartsAsIntegers[2], dateTimePartsAsIntegers[3], dateTimePartsAsIntegers[4], dateTimePartsAsIntegers[5], 0, time.UTC)
	clip, addClipErr := db.AddClip(ownerId, file.Filename, dateTime)
	if addClipErr != nil {
		println(addClipErr.Error())
		c.String(http.StatusInternalServerError, addClipErr.Error())
		return
	}

	c.IndentedJSON(http.StatusCreated, clip)
}

func DownloadClipById(c *gin.Context) {
	clipId, conversionErr := strconv.Atoi(c.Param("clipId"))
	if conversionErr != nil {
		c.String(http.StatusBadRequest, "invalid clip id provided: %s", c.Param("clipId"))
		return
	}
	clip, err := db.GetClipById(clipId)

	if err != nil {
		c.String(http.StatusBadRequest, "no clip found with id: %d", clipId)
		return
	}

	c.FileAttachment(config.GetOutputPath()+clip.Filename, clip.Filename)
}

func DownloadClipThumbnailById(c *gin.Context) {
	clipId, conversionErr := strconv.Atoi(c.Param("clipId"))
	if conversionErr != nil {
		c.String(http.StatusBadRequest, "invalid clip id provided: %s", c.Param("clipId"))
		return
	}
	clip, err := db.GetClipById(clipId)

	if err != nil {
		c.String(http.StatusBadRequest, "no clip found with id: %d", clipId)
		return
	}

	c.FileAttachment(config.GetThumbnailsPath()+clip.Filename+".png", clip.Filename+".png")
}
