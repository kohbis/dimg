package cmd

import (
	"testing"
)

func Test_commandExists(t *testing.T) {
	type args struct {
		cmd string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "return true if command exist",
			args: args{
				"ls",
			},
			want: true,
		},
		{
			name: "return false if command not exist",
			args: args{
				"may_not_exist",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := commandExists(tt.args.cmd); got != tt.want {
				t.Errorf("commandExists() = %v, want %v", got, tt.want)
			}
		})
	}
}
