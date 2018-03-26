package main

import (
	"database/sql"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const query string = `
SELECT
	coalesce(destination.nom_stand_dst, destination.nom_dst),
	t.code_graphic,
	horaire.hor
FROM
	parcours
		JOIN arret_sur_parcours USING ( ref_prc )
		JOIN destination USING ( ref_dst )
		JOIN arret a USING ( ref_art )
		JOIN ligne l USING (ref_lig)
		JOIN horaire USING ( ref_apr )
		LEFT JOIN arret t ON ( t.ref_art = parcours.art_dest )
		LEFT JOIN vehicule USING ( ref_crs )
WHERE
	a.num_art = $1 AND
	l.num_lig = $2 AND
	sens = $3 AND
	coalesce( horaire.hor + ( avance - retard ), hor) BETWEEN $4 AND
	CASE
	WHEN $4::time <= rezo_get_hdeb_exploit()::time
		THEN rezo_get_hdeb_exploit()::time
	ELSE
		'23:59:59'::time
	END
	AND (vehicule.actif = true OR vehicule.actif IS NULL)
ORDER BY hor
LIMIT $5
`

func NextDeparture(db *sql.DB, request SchedulesRequest) ([]ExpectedStopTime, error) {
	row, err := db.Query(query, request.Stop, request.Line, request.Way, request.Datetime, request.Count)
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