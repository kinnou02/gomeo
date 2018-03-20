package main

import (
	"net/http"
	"time"

	"database/sql"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const query string = `SELECT
    coalesce( destination.nom_stand_dst, destination.nom_dst ),
    t.code_graphic,
    horaire.hor
FROM
    parcours
        JOIN arret_sur_parcours USING ( ref_prc )
        JOIN destination USING ( ref_dst )
        JOIN arret a USING ( ref_art )
        JOIN horaire USING ( ref_apr )
        LEFT JOIN arret t ON ( t.ref_art = parcours.art_dest )
        LEFT JOIN vehicule USING ( ref_crs )
WHERE
    ref_prc IN ( SELECT ref_prc
                FROM parcours
                JOIN ligne l USING ( ref_lig )
                JOIN arret_sur_parcours USING ( ref_prc )
                JOIN arret a USING ( ref_art )
                WHERE
                    a.num_art = $1 AND
                    l.num_lig = $2 AND
                    sens = $3 ) AND
    a.num_art = $1  AND
    coalesce( horaire.hor + ( avance - retard ),hor) BETWEEN '09:00:00'::time AND '23:59:59'::time
    AND ( vehicule.actif = true OR vehicule.actif IS NULL )
ORDER BY hor
LIMIT 5
`

type Resp struct {
	List []Schedule
}
type Schedule struct {
	Destination string
	CodeGraphic string
	Time        time.Time
}

type SchedulesRequest struct {
	Line int    `form:"line" binding:"required"`
	Stop int    `form:"stop" binding:"required"`
	Way  string `form:"way" binding:"required"`
	//	Datetime time.Time `form:"datetime" time_format:"20060102T150405"`
}

func ScheduleHandler(db *sql.DB) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		var request SchedulesRequest
		if err := c.ShouldBindQuery(&request); err != nil {
			log.Errorf("FATAL: %+v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}
		row, err := db.Query(query, request.Stop, request.Line, request.Way)
		if err != nil {
			log.Errorf("FATAL: %+v\n", err)
			c.JSON(http.StatusInsufficientStorage, gin.H{"error": err})
			return
		}
		defer row.Close()
		var resp Resp
		for row.Next() {
			log.Debug("iter")
			var schedule Schedule
			err = row.Scan(&schedule.Destination, &schedule.CodeGraphic, &schedule.Time)
			resp.List = append(resp.List, schedule)
			if err != nil {
				log.Errorf("FATAL: %+v\n", err)
			}
		}
		err = row.Err()
		if err != nil {
			log.Errorf("FATAL: %+v\n", err)
		}
		log.Debug("finish")

		c.JSON(http.StatusOK, resp)
	}
	return gin.HandlerFunc(fn)
}
