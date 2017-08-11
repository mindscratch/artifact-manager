package core

import "log"

// Log logs a message using the given format and values.
func Log(format string, values ...interface{}) {
	log.Printf(format, values...)
}
