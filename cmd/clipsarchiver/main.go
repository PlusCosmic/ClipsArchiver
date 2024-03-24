package main

import (
	"ClipsArchiver/internal/config"
	"ClipsArchiver/internal/db"
	"github.com/gin-gonic/gin"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const logFileLocation = "clipsarchiver.log"
const error500String = "Something went wrong :("

var logger *slog.Logger

func main() {

	options := &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}

	file, err := os.OpenFile(logFileLocation, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to get log file handle: %s", err.Error())
	}

	var handler slog.Handler = slog.NewJSONHandler(file, options)
	logger = slog.New(handler)

	err = db.SetupDb(logger)
	if err != nil {
		log.Fatalf("Failed to setup database: %s", err.Error())
	}

	router := gin.Default()

	router.POST("/clips/upload/:ownerId", uploadClip)
	router.GET("/clips/:clipId", getClip)
	router.PUT("/clips/:clipId", updateClip)
	router.DELETE("/clips/:clipId", deleteClip)
	router.GET("/clips/filename/:filename", getClipByFilename)
	router.GET("/users", getAllUsers)
	router.GET("/tags", getAllTags)
	router.GET("/maps", getAllMaps)
	router.GET("/legends", getAllLegends)
	router.GET("/clips/queue", getClipsQueue)
	router.GET("/clips/queue/:clipId", getQueueEntryById)
	router.GET("/clips/date/:date", getClipsForDate) // YYYY-MM-DD
	router.GET("/clips/download/:clipId", downloadClipById)
	router.GET("/clips/download/thumbnail/:clipId", downloadClipThumbnailById)
	router.StaticFS("/clips/archive", http.Dir(config.GetOutputPath()))
	router.StaticFS("/resources", http.Dir(config.GetResourcesPath()))

	routerErr := router.Run()
	if routerErr != nil {
		log.Fatalf("Failed to start gin router: %s", routerErr)
	}
}

func uploadClip(c *gin.Context) {
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

func getClipByFilename(c *gin.Context) {
	clip, err := db.GetClipByFilename(c.Param("filename"))
	if err != nil {
		println(err.Error())
		c.String(http.StatusNotFound, "No clip found for filename")
		return
	}
	c.IndentedJSON(http.StatusOK, clip)
}

func getAllUsers(c *gin.Context) {
	users, err := db.GetAllUsers()
	if err != nil {
		c.String(http.StatusInternalServerError, error500String)
		return
	}
	c.IndentedJSON(http.StatusOK, users)
}

func getAllTags(c *gin.Context) {
	tags, err := db.GetAllTags()
	if err != nil {
		c.String(http.StatusInternalServerError, error500String)
		return
	}
	c.IndentedJSON(http.StatusOK, tags)
}

func getAllMaps(c *gin.Context) {
	gameMaps, err := db.GetAllMaps()
	if err != nil {
		c.String(http.StatusInternalServerError, error500String)
		return
	}
	c.IndentedJSON(http.StatusOK, gameMaps)
}

func getAllLegends(c *gin.Context) {
	legends, err := db.GetAllLegends()
	if err != nil {
		c.String(http.StatusInternalServerError, error500String)
		return
	}
	c.IndentedJSON(http.StatusOK, legends)
}

func getClipsQueue(c *gin.Context) {
	queueEntries, err := db.GetClipsQueue()
	if err != nil {
		c.String(http.StatusInternalServerError, error500String)
		return
	}
	c.IndentedJSON(http.StatusOK, queueEntries)
}

func getQueueEntryById(c *gin.Context) {
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
		println(err)
		c.String(http.StatusInternalServerError, error500String)
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

	c.FileAttachment(config.GetOutputPath()+clip.Filename, clip.Filename)
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

	c.FileAttachment(config.GetThumbnailsPath()+clip.Filename+".png", clip.Filename+".png")
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
		c.String(http.StatusInternalServerError, error500String)
		return
	}
	err = db.UpdateClipTags(existingClip, clip)
	if err != nil {
		c.String(http.StatusInternalServerError, error500String)
	}
	err = db.UpdateClip(clip)
	if err != nil {
		c.String(http.StatusInternalServerError, error500String)
	}
}

func deleteClip(c *gin.Context) {
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

func getClip(c *gin.Context) {
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
