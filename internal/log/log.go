package log

import (
	"log"
	"os"
	"strconv"
)

var enabled = false

func init() {
	if v, ok := os.LookupEnv("PUMACI_LOG"); ok {
		if lv, err := strconv.ParseBool(v); err == nil {
			enabled = lv
		}
	}
}

func Debug(f string, v ...any) {
	if enabled {
		log.Printf(f, v...)
	}
}

func Warn(f string, v ...any) {
	log.Printf(f, v...)
}
