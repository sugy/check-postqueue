package checkpostqueue

import (
	"reflect"
	"testing"
)

func TestConfig_loadConfig(t *testing.T) {
	type fields struct {
		Prefix        string
		PostqueuePath string
		Messages      []string
	}
	type args struct {
		configFile string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "check config",
			fields: fields{},
			args: args{
				configFile: "../testdata/config.toml",
			},
			wantErr: false,
		},
		{
			name:   "check error loading the file",
			fields: fields{},
			args: args{
				configFile: "../testdata/dummy.toml",
			},
			wantErr: true,
		},
		{
			name:   "check error decoding toml",
			fields: fields{},
			args: args{
				configFile: "../testdata/config_fail.toml",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{}
			if err := c.loadConfig(tt.args.configFile); (err != nil) != tt.wantErr {
				t.Errorf("Config.loadConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_generateConfig(t *testing.T) {
	tests := []struct {
		name       string
		wantOutput []string
	}{
		{
			name: "check generating config toml",
			wantOutput: []string{
				`# CheckPostqueue config file`,
				`# Path to postqueue command`,
				`PostqueuePath = "/usr/sbin/postqueue"`,
				``,
				`# Exclude messages`,
				`# Format: [ "<regex>", ... ]`,
				`ExcludeMessages = [`,
				//`  "Connection refused",`,
				//`  "Connection timed out",`,
				`  "Helo command rejected: Host not found",`,
				`  "type=MX: Host not found, try again",`,
				`  "Mailbox full",`,
				//`  "Network is unreachable",`,
				`  "No route to host",`,
				`  "The email account that you tried to reach is over quota",`,
				//`  "Relay access denied",`,
				`]`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{}
			if output := c.generateConfig(); !reflect.DeepEqual(output, tt.wantOutput) {
				t.Errorf("Config.generateConfig() output = %v, wantoutput = %v, DeepEqual = %v", output, tt.wantOutput, reflect.DeepEqual(output, tt.wantOutput))
			}
		})
	}
}
