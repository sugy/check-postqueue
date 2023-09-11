package checkpostqueue

import (
	"reflect"
	"testing"
)

func TestCheckPostqueueConfig_loadConfig(t *testing.T) {
	type fields struct {
		Prefix        string
		PostqueuePath string
		MsgCategories map[string]string
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
			c := &CheckPostqueueConfig{}
			if err := c.loadConfig(tt.args.configFile); (err != nil) != tt.wantErr {
				t.Errorf("CheckPostqueueConfig.loadConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckPostqueueConfig_generateConfig(t *testing.T) {
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
				`# Exclude message categories`,
				`# Format: <category> = "<regex>"`,
				`[ExcludeMsgCategories]`,
				//`  "Connection refused" = "Connection refused"`,
				//`  "Connection timeout" = "Connection timed out"`,
				`  "Helo command rejected" = "Helo command rejected: Host not found"`,
				`  "Host not found" = "type=MX: Host not found, try again"`,
				`  "Mailbox full" = "Mailbox full"`,
				//`  "Network is unreachable" = "Network is unreachable"`,
				`  "No route to host" = "No route to host"`,
				`  "Over quota" = "The email account that you tried to reach is over quota"`,
				//`  "Relay access denied" = "Relay access denied"`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CheckPostqueueConfig{}
			if output := c.generateConfig(); !reflect.DeepEqual(output, tt.wantOutput) {
				t.Errorf("CheckPostqueueConfig.generateConfig() output = %v, wantoutput = %v, DeepEqual = %v", output, tt.wantOutput, reflect.DeepEqual(output, tt.wantOutput))
			}
		})
	}
}
