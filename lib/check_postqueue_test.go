package checkpostqueue

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
	"testing"
)

// Helper function to read a file and return its contents as a string
func readFile(filePath string) (string, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("An error occurred while opening the file: %w", err)
	}
	defer file.Close()

	// Load the file's contents to a string
	contents, err := ioutil.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("An error occurred while loading the file: %w", err)
	}

	return string(contents), nil
}

func TestCheckPostqueue_AnalyzePostqueue(t *testing.T) {
	type fields struct {
		TestdataPath  string
		MsgCategories map[string]*regexp.Regexp
	}
	tests := []struct {
		name    string
		fields  fields
		want    map[string]float64
		wantErr bool
	}{
		{
			name: "check metrics",
			fields: fields{
				TestdataPath: "../testdata/postqueue_output.txt",
				MsgCategories: map[string]*regexp.Regexp{
					"Connection refused":    regexp.MustCompile(`Connection refused`),
					"Connection timeout":    regexp.MustCompile(`Connection timed out`),
					"Helo command rejected": regexp.MustCompile(`Helo command rejected: Host not found`),
					"Over quota":            regexp.MustCompile(`The email account that you tried to reach is over quota`),
				},
			},
			want: map[string]float64{
				"Connection_timeout":    4,
				"Connection_refused":    1,
				"Helo_command_rejected": 1,
				"Over_quota":            1,
				"queue":                 12,
			},
		},
		{
			name: "check metrics with empty msgCategories",
			fields: fields{
				TestdataPath:  "../testdata/postqueue_output.txt",
				MsgCategories: map[string]*regexp.Regexp{},
			},
			want: map[string]float64{
				"queue": 12,
			},
		},
		{
			// If there are no matching rows in msgCategories, the metric will be 0.
			name: "check metrics with no match msgCategories",
			fields: fields{
				TestdataPath: "../testdata/postqueue_output.txt",
				MsgCategories: map[string]*regexp.Regexp{
					"Connection timeout": regexp.MustCompile(`Connection timed out`),
					"Dummy":              regexp.MustCompile(`Dummy`),
				},
			},
			want: map[string]float64{
				"Connection_timeout": 4,
				"Dummy":              0,
				"queue":              12,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// テスト用の出力を読み込む
			output, err := readFile(tt.fields.TestdataPath)
			if err != nil {
				t.Errorf("CheckPostqueue.AnalyzePostqueue() error = %v", err)
				return
			}

			p := &CheckPostqueue{
				PostqueueOutput:      output,
				ExcludeMsgCategories: tt.fields.MsgCategories,
			}
			got, err := p.AnalyzePostqueue()
			if (err != nil) != tt.wantErr {

				t.Errorf("CheckPostqueue.AnalyzePostqueue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CheckPostqueue.AnalyzePostqueue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckPostqueue_loadConfig(t *testing.T) {
	type fields struct {
		PostqueuePath        string
		PostqueueArgs        []string
		PostqueueOutput      string
		ExcludeMsgCategories map[string]*regexp.Regexp
	}
	type args struct {
		configFile string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    *CheckPostqueue
	}{
		{
			name:   "check config",
			fields: fields{},
			args: args{
				configFile: "../testdata/config.toml",
			},
			wantErr: false,
			want: &CheckPostqueue{
				PostqueuePath: "/usr/bin/postqueue",
				ExcludeMsgCategories: map[string]*regexp.Regexp{
					//"Connection refused":     regexp.MustCompile(`Connection refused`),
					//"Connection timeout":     regexp.MustCompile(`Connection timed out`),
					"Helo command rejected": regexp.MustCompile(`Helo command rejected: Host not found`),
					"Host not found":        regexp.MustCompile(`type=MX: Host not found, try again`),
					"Mailbox full":          regexp.MustCompile(`Mailbox full`),
					//"Network is unreachable": regexp.MustCompile(`Network is unreachable`),
					"No route to host": regexp.MustCompile(`No route to host`),
					"Over quota":       regexp.MustCompile(`The email account that you tried to reach is over quota`),
					//"Relay access denied":    regexp.MustCompile(`Relay access denied`),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &CheckPostqueue{}
			err := p.loadConfig(tt.args.configFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPostqueue.loadConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(p, tt.want) {
				t.Errorf("CheckPostqueue.loadConfig() = %v, want %v", p, tt.want)
			}
		})
	}
}
