package checkpostqueue

import (
	"fmt"
	"runtime"
	"time"
)

var (
	version = "0.0.0-dev"
	commit  = "none"
	date    = "unknown"
)

func showVersion() {
	fmt.Printf("version %s (rev %s, %s) [%s %s %s] \n",
		version, commit, UTCToJST(date), runtime.GOOS, runtime.GOARCH, runtime.Version())
}

// UTCToJST converts UTC to JST
func UTCToJST(utc string) string {
	d, err := time.Parse("2006-01-02T15:04:05Z", utc) //Parse RFC3339
	if err != nil {
		return utc
	}
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	return d.In(jst).Format(time.RFC3339)
}
