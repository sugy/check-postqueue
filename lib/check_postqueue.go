package checkpostqueue

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	flags "github.com/jessevdk/go-flags"
	"github.com/mackerelio/checkers"
	log "github.com/sirupsen/logrus"
)

// Do the plugin
func Do() {
	customFmt := new(log.TextFormatter)
	customFmt.TimestampFormat = "2006-01-02 15:04:05"
	customFmt.FullTimestamp = true
	log.SetFormatter(customFmt)
	log.SetOutput(os.Stdout)

	ckr := run(os.Args[1:])
	ckr.Name = "Postqueue"
	ckr.Exit()
}

type monitor struct {
	warning  int64
	critical int64
}

func (m monitor) hasWarning() bool {
	return m.warning != 0
}

func (m monitor) checkWarning(queue int64) bool {
	return (m.hasWarning() && m.warning < queue)
}

func (m monitor) hasCritical() bool {
	return m.critical != 0
}

func (m monitor) checkCritical(queue int64) bool {
	return (m.hasCritical() && m.critical < queue)
}

func newMonitor(warning, critical int64) *monitor {
	return &monitor{
		warning:  warning,
		critical: critical,
	}
}

var opts struct {
	Warning  int64 `short:"w" long:"warning" default:"100" description:"number of messages in queue to generate warning"`
	Critical int64 `short:"c" long:"critical" default:"200" description:"number of messages in queue to generate critical alert ( w < c )"`
	Version  bool  `long:"version" description:"Show version"`
	Debug    bool  `long:"debug" description:"Debug mode"`
}

func run(args []string) *checkers.Checker {
	_, err := flags.ParseArgs(&opts, args)
	if err != nil {
		os.Exit(1)
	}

	if opts.Debug {
		log.SetLevel(log.DebugLevel)
	}

	if opts.Version {
		showVersion()
		os.Exit(0)
	}

	var queue int64
	queueStr := "0"
	monitor := newMonitor(opts.Warning, opts.Critical)

	result := checkers.OK

	out, err := exec.Command("mailq").Output()

	if err != nil {
		return checkers.Unknown(err.Error())
	}

	outs := strings.Split(string(out), "\n")
	line := outs[len(outs)-2]

	re := regexp.MustCompile(`-- \d+ Kbytes in (\d+) Requests.`)
	if re.MatchString(line) {
		queueStr = re.ReplaceAllString(line, "$1")
		queue, err = strconv.ParseInt(queueStr, 10, 64)
	}

	if monitor.checkWarning(queue) {
		result = checkers.WARNING
	}

	if monitor.checkCritical(queue) {
		result = checkers.CRITICAL
	}

	msg := fmt.Sprintf(queueStr)
	return checkers.NewChecker(result, msg)
}
