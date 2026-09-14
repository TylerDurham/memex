// Package walkers
package walkers

import "time"

type Document struct {
	AbsPath     string
	Application string
	ModTime     time.Time
	RelPath     string
	Size        int64
	URI         string
}
