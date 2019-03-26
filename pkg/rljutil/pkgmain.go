// Package rljutil is for stupid utilities
package rljutil

import (
	"log"

	"github.com/pkg/errors"
)

// FatalIf rasies fatal error if err is not nil
func FatalIf(err error, format string, args ...interface{}) {
	if err != nil {
		log.Fatal(errors.Wrapf(err, format, args...))
	}
}
