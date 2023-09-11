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
	PostqueuePath   string
	PostqueueArgs   []string
	PostqueueOutput string
	ExcludeMessages []*regexp.Regexp
}

// Analyze Postfix postqueue
func (p *CheckPostqueue) AnalyzePostqueue() ([]int64, error) {
	if p.PostqueueOutput == "" {
		output, err := p.runPostQueueCommand()
		if err != nil {
			return nil, err
		}
		log.Debug("Fetch (output): ", fmt.Sprintf("'%v'", output))
		p.PostqueueOutput = output
	}

	// Initialize counts
	counts := make([]int64, len(p.ExcludeMessages)+1)

	// Read and count output entries
	scanner := bufio.NewScanner(strings.NewReader(p.PostqueueOutput))
	for scanner.Scan() {
		line := scanner.Text()
		for i := range p.ExcludeMessages {
			if p.ExcludeMessages[i].MatchString(line) {
				// log.Debug("AnalyzePostqueue (line): ", fmt.Sprintf("'%v'", line))
				counts[i] = counts[i] + 1
				break
			}
		}
		// line の先頭が 10桁以上の16進数であれば、それはキューIDとみなす
		if len(line) >= 10 && regexp.MustCompile(`^[0-9A-F]{10,12}[*]{0,1}\s+`).MatchString(line) {
			counts[len(counts)-1] = counts[len(counts)-1] + 1
		}
	}
	log.Debug("AnalyzePostqueue (counts): ", fmt.Sprintf("'%v'", counts))
	return counts, nil
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
	if c.ExcludeMessages != nil {
		p.ExcludeMessages = getMessages(c.ExcludeMessages)
	}

	return nil
}

// getMsgCategories returns the map of message categories
func getMessages(m []string) []*regexp.Regexp {
	messages := make([]*regexp.Regexp, len(m))
	for i := range m {
		messages[i] = regexp.MustCompile(m[i])
	}
	return messages
}

func (p *CheckPostqueue) validate() error {
	if p.PostqueuePath == "" {
		return fmt.Errorf("postqueue path is required")
	}
	if len(p.ExcludeMessages) == 0 {
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
	if p.ExcludeMessages == nil {
		p.ExcludeMessages = getMessages(getDefaultMessages())
	}

	log.Debug("Do (p): ", fmt.Sprintf("'%v'", p))

	err = p.validate()
	if err != nil {
		log.Errorf("Failed to validate config: %s", err)
		os.Exit(1)
	}

	var counts []int64
	counts, err = p.AnalyzePostqueue()
	if err != nil {
		log.Errorf("Failed to Analyze Postqueue: %s", err)
		os.Exit(1)
	}
	log.Debug("counts: ", fmt.Sprintf("'%v'", counts))

	queue := int64(counts[len(counts)-1])
	monitor := newMonitor(opts.Warning, opts.Critical)
	result := checkers.OK

	exclude := int64(0)
	for i := range p.ExcludeMessages {
		exclude = exclude + int64(counts[i])
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
