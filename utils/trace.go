package utils

import (
	"log"
	"time"
)

//Trace record function run information
func Trace(msg string) func() {
	t1 := time.Now()
	log.Println(" Enter [ " + msg + " ]")
	return func() {
		log.Println(" Leave [ "+msg+" ], takes times:", time.Since(t1))
	}
}
