package checkpostqueue

import (
	"fmt"
	"io/ioutil"
	"sort"

	"github.com/BurntSushi/toml"
	log "github.com/sirupsen/logrus"
)

// CheckPostqueueConfig is the configuration file format
type CheckPostqueueConfig struct {
	PostqueuePath        string
	ExcludeMsgCategories map[string]string
}

// loadConfig loads the plugin configuration file
func (c *CheckPostqueueConfig) loadConfig(configFile string) error {
	contents, err := ioutil.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("an error occurred while loading the file: %w", err)
	}

	// Perse TOML in contents
	if _, err := toml.Decode(string(contents), &c); err != nil {
		return fmt.Errorf("an error occurred while decoding TOML format file: %w", err)
	}

	return nil
}

// Generate config file template
func (c *CheckPostqueueConfig) generateConfig() []string {
	c.PostqueuePath = "/usr/sbin/postqueue"

	c.ExcludeMsgCategories = getDefaultMsgCategories()
	keys := c.getMsgCategoriesKeys()
	sort.Strings(keys)
	log.Debug("generateConfig: MsgCategories keys: ", keys)

	var result []string
	result = append(result, `# CheckPostqueue config file`)

	// Output config file template
	result = append(result, `# Path to postqueue command`)
	result = append(result, `PostqueuePath = "`+c.PostqueuePath+`"`)
	result = append(result, ``)

	result = append(result, `# Exclude message categories`)
	result = append(result, `# Format: <category> = "<regex>"`)
	result = append(result, `[ExcludeMsgCategories]`)
	for k := range keys {
		result = append(result, `  "`+keys[k]+`" = "`+c.ExcludeMsgCategories[keys[k]]+`"`)
	}

	return result
}

// Get MsgCategories keys
func (c *CheckPostqueueConfig) getMsgCategoriesKeys() []string {
	keys := make([]string, 0, len(c.ExcludeMsgCategories))
	for k := range c.ExcludeMsgCategories {
		keys = append(keys, k)
	}
	return keys
}

// Set default MsgCategories
func getDefaultMsgCategories() map[string]string {
	return map[string]string{
		//"Connection refused":     "Connection refused",
		//"Connection timeout":     "Connection timed out",
		"Helo command rejected": "Helo command rejected: Host not found",
		"Host not found":        "type=MX: Host not found, try again",
		"Mailbox full":          "Mailbox full",
		//"Network is unreachable": "Network is unreachable",
		"No route to host": "No route to host",
		"Over quota":       "The email account that you tried to reach is over quota",
		//"Relay access denied":    "Relay access denied",
		// Add more log categories with corresponding regular expressions
	}
}
