package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/bradstell/future"
	"github.com/gin-gonic/gin"
)

type AppointmentPostBody struct {
	UserID   int       `json:"user_id"`
	StartsAt time.Time `json:"starts_at"`
}

func errsToJson(errs []error) string {
	// we have some errors lets return them back to the client
	errStrs := make([]string, 0, len(errs))
	for _, e := range errs {
		errStrs = append(errStrs, e.Error())
	}
	bytes, _ := json.Marshal(errStrs)
	return string(bytes)
}

func main() {
	// Creates a gin router with default middleware
	r := gin.Default()

	// get a list of available appointment times for a trainer between two dates
	// params: trainer_id, starts_at, ends_at
	r.GET("trainers/:trainer_id/appointments", func(c *gin.Context) {
		// extract query params as strings
		trainerIDStr := c.Param("trainer_id")
		startsAtISO := c.Query("starts_at")
		endsAtISO := c.Query("ends_at")

		var errs []error

		// parse values into go types
		trainerID, err := strconv.Atoi(trainerIDStr)
		if err != nil {
			errs = append(errs, fmt.Errorf("trainer ID '%s' not a valid integer", trainerIDStr))
		}

		startDate, err := time.Parse(time.RFC3339, startsAtISO)
		if err != nil {
			errs = append(errs, fmt.Errorf("starts_at value of '%s' not in valid RFC3339 format", startsAtISO))
		}

		endDate, err := time.Parse(time.RFC3339, endsAtISO)
		if err != nil {
			errs = append(errs, fmt.Errorf("ends_at value of '%s' not in valid RFC3339 format", endsAtISO))
		}

		if len(errs) > 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"errors": errsToJson(errs),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": future.GetAvailableAppointmentsFor(trainerID, startDate, endDate),
		})
	})

	// create a new apt with a trainer
	r.POST("trainers/:trainer_id/appointments", func(c *gin.Context) {
		// extract query params as strings
		trainerIDStr := c.Param("trainer_id")
		var postData AppointmentPostBody

		if err := c.ShouldBindJSON(&postData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// parse values into go types
		trainerID, err := strconv.Atoi(trainerIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("trainer ID '%s' not a valid int", trainerIDStr),
			})
			return
		}

		appointment, err := future.BookAppointmentFor(trainerID, postData.UserID, postData.StartsAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": appointment,
		})
	})

	// get a list of scheduled appointments for a trainer
	// TODO whats a better way to structure this path?
	r.GET("trainers/:trainer_id/appointments/scheduled", func(c *gin.Context) {
		trainerIDStr := c.Param("trainer_id")
		trainerID, err := strconv.Atoi(trainerIDStr)
		if err != nil {
			// handle error response to client
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("trainer ID '%s' not a valid int", trainerIDStr),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": future.GetScheduledAppointmentsFor(trainerID),
		})
	})

	// start server
	r.Run(":3000")
}
