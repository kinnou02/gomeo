package main

import (
	"flag"
	"os"
	"time"

	"database/sql"

	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func setupRouter() *gin.Engine {
	r := gin.New()
	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	r.Use(gin.Recovery())

	r.Use(ginrus.Ginrus(logrus.StandardLogger(), time.RFC3339, false))
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	return r
}

func init_log(logjson bool) {
	// Log as JSON instead of the default ASCII formatter.
	if logjson {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	logrus.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	logrus.SetLevel(logrus.DebugLevel)
}

var (
	httpDurations = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_durations_seconds",
		Help:    "http request latency distributions.",
		Buckets: prometheus.ExponentialBuckets(0.001, 1.5, 25),
	},
		[]string{"handler"},
	)
)

func init() {
	prometheus.MustRegister(httpDurations)
}

func Instrument(handler string) gin.HandlerFunc {
	return func(c *gin.Context) {
		observer := httpDurations.With(prometheus.Labels{"handler": handler})
		begin := time.Now()
		c.Next()
		observer.Observe(time.Since(begin).Seconds())
	}
}

func main() {
	listen := flag.String("listen", ":8080", "[IP]:PORT to listen")
	logjson := flag.Bool("logjson", false, "enable json logging")
	connStr := flag.String("connstr", "host=localhost user=timeo password=timeo dbname=timeo sslmode=disable", "")
	flag.Parse()
	init_log(*logjson)
	db, err := sql.Open("postgres", *connStr)
	if err != nil {
		logrus.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		logrus.Fatal(err)
	}

	r := setupRouter()
	r.GET("/schedules", Instrument("schedules"), ScheduleHandler(db))
	// Listen and Server in 0.0.0.0:8080
	err = r.Run(*listen)
	if err != nil {
		logrus.Errorf("failure to start: %+v", err)
	}
}
