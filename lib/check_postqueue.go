package checkpostqueue

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	flags "github.com/jessevdk/go-flags"
	"github.com/mackerelio/checkers"
	log "github.com/sirupsen/logrus"
)

type CheckPostqueue struct {
	PostqueuePath        string
	PostqueueArgs        []string
	PostqueueOutput      string
	ExcludeMsgCategories map[string]*regexp.Regexp
}

// Analyze Postfix postqueue
func (p *CheckPostqueue) AnalyzePostqueue() (map[string]float64, error) {
	if p.PostqueueOutput == "" {
		output, err := p.runPostQueueCommand()
		if err != nil {
			return nil, err
		}
		log.Debug("Fetch (output): ", fmt.Sprintf("'%v'", output))
		p.PostqueueOutput = output
	}

	// Initialize metric map
	metrics := make(map[string]float64)
	// Initialize metric map for Message categories
	for category := range p.ExcludeMsgCategories {
		name := strings.Replace(category, " ", "_", -1)
		metrics[name] = 0
	}

	// Read and classify output entries
	scanner := bufio.NewScanner(strings.NewReader(p.PostqueueOutput))
	for scanner.Scan() {
		line := scanner.Text()
		for category, regex := range p.ExcludeMsgCategories {
			if regex.MatchString(line) {
				// log.Debug("AnalyzePostqueue (line): ", fmt.Sprintf("'%v'", line))
				name := strings.Replace(category, " ", "_", -1)
				metrics[name] = metrics[name] + 1
				// not break here, because one line may match multiple categories
			}
		}
		// line の先頭が 10桁以上の16進数であれば、それはキューIDとみなす
		if len(line) >= 10 && regexp.MustCompile(`^[0-9A-F]{10,12}[*]{0,1}\s+`).MatchString(line) {
			metrics["queue"] = metrics["queue"] + 1
		}
	}
	log.Debug("AnalyzePostqueue (metrics): ", fmt.Sprintf("'%v'", metrics))
	return metrics, nil
}

// runPostQueueCommand executes the "postqueue -p" command and returns its output
func (p *CheckPostqueue) runPostQueueCommand() (string, error) {
	log.Debug(fmt.Sprintf("command: %v %v", p.PostqueuePath, strings.Join(p.PostqueueArgs, " ")))
	cmd := exec.Command(p.PostqueuePath, p.PostqueueArgs...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := cmd.ProcessState.ExitCode()

	if err != nil {
		return "", fmt.Errorf(fmt.Sprintf("failed to execute postqueue command. exitCode: %d, Stdout: '%s', Stderr: '%s'\n", exitCode, stdout.String(), stderr.String()))
	}
	return stdout.String(), err
}

// loadConfig loads config file
func (p *CheckPostqueue) loadConfig(configFile string) error {
	c := &CheckPostqueueConfig{}
	// Load config file
	err := c.loadConfig(configFile)
	if err != nil {
		return err
	}

	// Set config file values
	if c.PostqueuePath != "" {
		p.PostqueuePath = c.PostqueuePath
	}

	// Set config file values for Message categories
	if c.ExcludeMsgCategories != nil {
		p.ExcludeMsgCategories = getMsgCategories(c.ExcludeMsgCategories)
	}

	return nil
}

// getMsgCategories returns the map of message categories
func getMsgCategories(m map[string]string) map[string]*regexp.Regexp {
	msgCategories := make(map[string]*regexp.Regexp)
	for category, regex := range m {
		if category != "" && regex != "" {
			msgCategories[category] = regexp.MustCompile(regex)
		}
	}
	return msgCategories
}

func (p *CheckPostqueue) validate() error {
	if p.PostqueuePath == "" {
		return fmt.Errorf("postqueue path is required")
	}
	if len(p.ExcludeMsgCategories) == 0 {
		return fmt.Errorf("message categories is required")
	}
	return nil
}

// Generate config file template
func generateConfig() {
	c := &CheckPostqueueConfig{}
	output := c.generateConfig()
	for _, line := range output {
		fmt.Println(line)
	}
}

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
	Warning  int64  `short:"w" long:"warning" default:"100" description:"number of messages in queue to generate warning"`
	Critical int64  `short:"c" long:"critical" default:"200" description:"number of messages in queue to generate critical alert ( w < c )"`
	Version  bool   `long:"version" description:"Show version"`
	Debug    bool   `long:"debug" description:"Debug mode"`
	Path     string `long:"path" description:"Path to postqueue command" default:"/usr/sbin/postqueue"`
	Config   string `long:"config" description:"Path to TOML format config file"`
	GenConf  bool   `long:"generate-config" description:"Generate config file template"`
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

	if opts.GenConf {
		generateConfig()
		os.Exit(0)
	}

	p := &CheckPostqueue{}
	p.PostqueueArgs = []string{"-p"}
	if opts.Path != "" {
		p.PostqueuePath = opts.Path
	}

	if opts.Config != "" {
		err := p.loadConfig(opts.Config)
		if err != nil {
			log.Errorf("Failed to load config file: %s", err)
			os.Exit(1)
		}
	}

	// Set default values for Message categories
	if p.ExcludeMsgCategories == nil {
		p.ExcludeMsgCategories = getMsgCategories(getDefaultMsgCategories())
	}

	log.Debug("Do (p): ", fmt.Sprintf("'%v'", p))

	err = p.validate()
	if err != nil {
		log.Errorf("Failed to validate config: %s", err)
		os.Exit(1)
	}

	var metrics map[string]float64
	metrics, err = p.AnalyzePostqueue()
	if err != nil {
		log.Errorf("Failed to Analyze Postqueue: %s", err)
		os.Exit(1)
	}
	log.Debug("metrics: ", fmt.Sprintf("'%v'", metrics))

	queue := int64(metrics["queue"])
	monitor := newMonitor(opts.Warning, opts.Critical)
	result := checkers.OK

	exclude := int64(0)
	for category := range p.ExcludeMsgCategories {
		name := strings.Replace(category, " ", "_", -1)
		exclude = exclude + int64(metrics[name])
	}

	count := int64(0)
	count = queue - exclude

	if monitor.checkWarning(count) {
		result = checkers.WARNING
	}

	if monitor.checkCritical(count) {
		result = checkers.CRITICAL
	}

	msg := fmt.Sprintf("count: %d (queue: %d, exclude: %d)", count, queue, exclude)
	return checkers.NewChecker(result, msg)
}
