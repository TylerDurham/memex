module github.com/tylersnork/memex

go 1.22.2

require (
	github.com/fsnotify/fsnotify v1.7.0
	github.com/mattn/go-sqlite3 v1.14.49
)

require golang.org/x/sys v0.4.0 // indirect

replace golang.org/x/sys => github.com/golang/sys v0.4.0
