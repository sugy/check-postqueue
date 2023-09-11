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
		TestdataPath string
		Messages     []*regexp.Regexp
	}
	tests := []struct {
		name    string
		fields  fields
		want    []int64
		wantErr bool
	}{
		{
			name: "check counts",
			fields: fields{
				TestdataPath: "../testdata/postqueue_output.txt",
				Messages: []*regexp.Regexp{
					regexp.MustCompile(`Connection refused`),
					regexp.MustCompile(`Connection timed out`),
					regexp.MustCompile(`Helo command rejected: Host not found`),
					regexp.MustCompile(`The email account that you tried to reach is over quota`),
				},
			},
			want: []int64{
				1,
				4,
				1,
				1,
				12,
			},
		},
		{
			name: "check counts with empty exclude messages",
			fields: fields{
				TestdataPath: "../testdata/postqueue_output.txt",
				Messages:     []*regexp.Regexp{},
			},
			want: []int64{
				12,
			},
		},
		{
			// If there are no matching rows in messages, the count will be 0.
			name: "check counts with no match exclude messages",
			fields: fields{
				TestdataPath: "../testdata/postqueue_output.txt",
				Messages: []*regexp.Regexp{
					regexp.MustCompile(`Connection timed out`),
					regexp.MustCompile(`Dummy`),
				},
			},
			want: []int64{
				4,
				0,
				12,
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
				PostqueueOutput: output,
				ExcludeMessages: tt.fields.Messages,
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
		PostqueuePath   string
		PostqueueArgs   []string
		PostqueueOutput string
		ExcludeMessages []*regexp.Regexp
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
				ExcludeMessages: []*regexp.Regexp{
					//regexp.MustCompile(`Connection refused`),
					//regexp.MustCompile(`Connection timed out`),
					regexp.MustCompile(`Helo command rejected: Host not found`),
					regexp.MustCompile(`type=MX: Host not found, try again`),
					regexp.MustCompile(`Mailbox full`),
					//regexp.MustCompile(`Network is unreachable`),
					regexp.MustCompile(`No route to host`),
					regexp.MustCompile(`The email account that you tried to reach is over quota`),
					//regexp.MustCompile(`Relay access denied`),
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
