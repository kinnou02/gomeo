package main

import (
	"database/sql"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

const query string = `
SELECT
	coalesce(destination.nom_stand_dst, destination.nom_dst),
	t.code_graphic,
	h.hor
FROM
	parcours
		JOIN arret_sur_parcours USING ( ref_prc )
		JOIN destination USING ( ref_dst )
		JOIN arret a USING ( ref_art )
		JOIN ligne l USING (ref_lig)
		JOIN horaire h USING ( ref_apr )
		LEFT JOIN arret t ON ( t.ref_art = parcours.art_dest )
		LEFT JOIN vehicule v USING ( ref_crs )
WHERE
	a.num_art = $1 AND
	l.num_lig = $2 AND
	sens = $3 AND
	coalesce( h.hor + ( v.avance - v.retard ), h.hor) BETWEEN $4 AND
	CASE
	WHEN $4::time <= rezo_get_hdeb_exploit()::time
		THEN rezo_get_hdeb_exploit()::time
	ELSE
		'23:59:59'::time
	END
	AND (v.actif = true OR v.actif IS NULL)
ORDER BY h.hor
LIMIT $5
`

var (
	sqlDurations = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "sql_durations_seconds",
		Help:    "sql request latency distributions.",
		Buckets: prometheus.ExponentialBuckets(0.001, 1.5, 25),
	})
)

func init() {
	prometheus.MustRegister(sqlDurations)
}

func NextDeparture(db *sql.DB, request SchedulesRequest) ([]ExpectedStopTime, error) {
	begin := time.Now()
	row, err := db.Query(query, request.Stop, request.Line, request.Way, request.Datetime, request.Count)
	sqlDurations.Observe(time.Since(begin).Seconds())
	if err != nil {
		return nil, errors.Wrap(err, "query failed")
	}
	defer func() {
		if err = row.Close(); err != nil {
			logrus.Error(err)
		}
	}()

	result := make([]ExpectedStopTime, 0, request.Count)
	for row.Next() {
		var schedule ExpectedStopTime
		err = row.Scan(&schedule.Destination, &schedule.TerminusSAECode, &schedule.NextStop)
		if err != nil {
			return nil, errors.Wrap(err, "Scan failed")
		}
		result = append(result, schedule)
	}

	err = row.Err()
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	return result, nil
}
