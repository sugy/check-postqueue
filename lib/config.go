package checkpostqueue

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// CheckPostqueueConfig is the configuration file format
type CheckPostqueueConfig struct {
	PostqueuePath   string
	ExcludeMessages []string
}

// loadConfig loads the plugin configuration file
func (c *CheckPostqueueConfig) loadConfig(configFile string) error {
	contents, err := os.ReadFile(configFile)
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

	c.ExcludeMessages = getDefaultMessages()

	var result []string
	result = append(result, `# CheckPostqueue config file`)

	// Output config file template
	result = append(result, `# Path to postqueue command`)
	result = append(result, `PostqueuePath = "`+c.PostqueuePath+`"`)
	result = append(result, ``)

	result = append(result, `# Exclude messages`)
	result = append(result, `# Format: [ "<regex>", ... ]`)
	result = append(result, `ExcludeMessages = [`)
	for i := range c.ExcludeMessages {
		result = append(result, `  "`+c.ExcludeMessages[i]+`",`)
	}
	result = append(result, `]`)

	return result
}

// Set default Messages
func getDefaultMessages() []string {
	return []string{
		//"Connection refused",
		//"Connection timed out",
		"Helo command rejected: Host not found",
		"type=MX: Host not found, try again",
		"Mailbox full",
		//"Network is unreachable",
		"No route to host",
		"The email account that you tried to reach is over quota",
		//"Relay access denied",
		// Add more log categories with corresponding regular expressions
	}
}
