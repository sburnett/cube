/*
	Package cube periodically sends expvars to a Cube Collector. This is handy
	if you want to create dashboards for visualizing statistics about
	long-running programs.

	You must first install and run a Cube collector. See
	http://square.github.com/cube/

	This module is plug-and-play. Example usage:

		package main

		import (
			"expvar"
			"flags"
			"github.com/sburnett/cube"
		)

		func main() {
			flags.Parse()  // You must parse flags before starting the exporter.
			go cube.Run("myevents")  // Runs forever, so run it in a goroutine.

			// Now create and use expvars.
		}

*/
package cube

import (
	"bytes"
	"expvar"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

var flagCollectorHost string
var flagCollectorPort int
var flagExportInterval string
var flagExportToCube bool

func init() {
	flag.StringVar(&flagCollectorHost, "cube_collector_host", "localhost", "Export variables to this Cube collector.")
	flag.IntVar(&flagCollectorPort, "cube_collector_port", 1080, "Use this port when connecting to the Cube collector.")
	flag.StringVar(&flagExportInterval, "cube_export_interval", "10s", "Export variables to Cube once every interval.")
	flag.BoolVar(&flagExportToCube, "cube_export", true, "Whether or not to export variables to Cube.")
}

// Periodically export variables from expvar to a Cube collector. This function
// never exits under normal circumstances, so you probably want to run it in a
// goroutine.
//
// You can control the collector hostname and port and how often we export to
// Cube using the cube_collector_host, cube_collector_port and
// cube_export_interval flags.
func Run(collectionType string) {
	if !flagExportToCube {
		return
	}

	putUrl := fmt.Sprintf("http://%s:%d/1.0/event/put", flagCollectorHost, flagCollectorPort)
	log.Printf("Exporting expvars to %s with event type %s", putUrl, collectionType)

	interval, err := time.ParseDuration(flagExportInterval)
	if err != nil {
		log.Fatalf("Error parsing duration %v: %v", flagExportInterval, err)
	}
	log.Printf("Exporting variables every %v", interval)

	exportCounter := expvar.NewInt("CubeExports")
	for now := range time.Tick(interval) {
		if err := ExportVariablesWithTimestamp(collectionType, putUrl, now); err != nil {
			log.Printf("Error exporting variables for %v", now)
		}
		exportCounter.Add(1)
	}
}

// Export expvars to Cube right now. Use the current system time as the timestamp
// for the submitted event. This function sends variables once and returns.
//
// You shouldn't need this function under normal circumstances. Use Run()
// instead.
func ExportVariables(collectionType string, putUrl string) error {
	return ExportVariablesWithTimestamp(collectionType, putUrl, time.Now())
}

// Export expvars to Cube right now. Use the provided timestamp for the
// submitted event. This function sends variables once and returns.
//
// You shouldn't need this function under normal circumstances. Use Run()
// instead.
func ExportVariablesWithTimestamp(collectionType string, putUrl string, timestamp time.Time) error {
	variables := make([]string, 0)
	expvar.Do(func(entry expvar.KeyValue) {
		variables = append(variables, fmt.Sprintf("%q: %s", entry.Key, entry.Value))
	})
	request := fmt.Sprintf(
		`[
		{
			"type": "%s",
			"time": "%s",
			"data": { %s }
		}
		]`,
		collectionType,
		timestamp.Format(time.ANSIC),
		strings.Join(variables, ","))

	response, err := http.Post(putUrl, "application/json", bytes.NewBufferString(request))
	if err != nil {
		log.Printf("Error POSTing events to Cube collector: %v", err)
		log.Printf("The request we tried to post: %v", request)
		return err
	}
	defer response.Body.Close()
	return nil
}
