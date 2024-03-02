package main

import (
	"ClipsArchiver/internal/db"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const inputPath = "/Uploads/"
const outputPath = "/Clips/"
const thumbnailsPath = "/Clips/Thumbnails/"
const storePath = "/Volumes/Big Store/TheArchive"

func main() {

	err := db.SetupDb()
	if err != nil {
		return
	}
	// Create Gin router
	router := gin.Default()

	// Register Routes
	router.POST("/clips/upload/:ownerId", uploadClip)
	router.PUT("/clips/:clipId", updateClip)
	router.GET("/users", getAllUsers)
	router.GET("/clips/queue", getClipsQueue)
	// YYYY-MM-DD
	router.GET("/clips/date/:date", getClipsForDate)
	router.GET("/clips/download/:clipId", downloadClipById)
	router.GET("/clips/download/thumbnail/:clipId", downloadClipThumbnailById)

	// Start the server
	routerErr := router.Run()
	if routerErr != nil {
		log.Fatal(routerErr)
		return
	}
}

func uploadClip(c *gin.Context) {
	ownerId, conversionErr := strconv.Atoi(c.Param("ownerId"))
	if conversionErr != nil {
		c.String(http.StatusBadRequest, "invalid owner id provided: %s", c.Param("ownerId"))
	}
	// Single file
	file, err := c.FormFile("file")
	log.Println(file.Filename)

	if err != nil {
		c.String(http.StatusBadRequest, "get form err: %s", err.Error())
		return
	}

	filename := storePath + inputPath + filepath.Base(file.Filename)
	if err := c.SaveUploadedFile(file, filename); err != nil {
		c.String(http.StatusBadRequest, "upload file err: %s", err.Error())
		return
	}

	addClipErr := db.AddClip(ownerId, file.Filename)
	if addClipErr != nil {
		c.String(http.StatusInternalServerError, addClipErr.Error())
		return
	}

	c.String(http.StatusOK, fmt.Sprintf("'%s' uploaded!", file.Filename))
}

func getAllUsers(c *gin.Context) {
	users, err := db.GetAllUsers()
	if err != nil {
		c.String(http.StatusInternalServerError, "Something went wrong :(")
		return
	}
	c.IndentedJSON(http.StatusOK, users)
}

func getClipsQueue(c *gin.Context) {
	queueEntries, err := db.GetClipsQueue()
	if err != nil {
		c.String(http.StatusInternalServerError, "Something went wrong :(")
		return
	}
	c.IndentedJSON(http.StatusOK, queueEntries)
}

func getClipsForDate(c *gin.Context) {
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
		c.String(http.StatusInternalServerError, "Something went wrong :(")
	}

	c.IndentedJSON(http.StatusOK, clips)
}

func downloadClipById(c *gin.Context) {
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

	c.FileAttachment(storePath+outputPath+clip.Filename, clip.Filename)
}

func downloadClipThumbnailById(c *gin.Context) {
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

	c.FileAttachment(storePath+thumbnailsPath+clip.Filename+".png", clip.Filename+".png")
}

func updateClip(c *gin.Context) {
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
		c.String(http.StatusInternalServerError, "Something went wrong :(")
		return
	}
	err = db.UpdateClip(existingClip, clip)
	if err != nil {
		c.String(http.StatusInternalServerError, "Something went wrong :(")
	}
}
