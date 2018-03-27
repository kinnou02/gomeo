package main

import (
	"net/http"

	"database/sql"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type SchedulesRequest struct {
	Line     int       `form:"line" binding:"required"`
	Stop     int       `form:"stop" binding:"required"`
	Way      string    `form:"way" binding:"required,len=1,containsany=AR"`
	Datetime time.Time `form:"datetime" time_format:"150405"`
	Count    int       `form:"count" `
}

func ScheduleHandler(db *sql.DB) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		var request SchedulesRequest
		if err := c.ShouldBindQuery(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}
		if request.Count == 0 {
			request.Count = 5
		}
		schedules, err := NextDeparture(db, request)
		if err != nil {
			log.Errorf("%+v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		r := Response{
			StopTimesResponse: []StopTimesResponse{{
				NextStopTimesMessage: NextStopTimesMessage{
					NextExpectedStopTime: schedules,
				},
			}},
		}
		c.JSON(http.StatusOK, r)
	}
	return gin.HandlerFunc(fn)
}
