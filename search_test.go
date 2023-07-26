package main

import (
	"reflect"
	"testing"
)

func TestPreParseQuery(t *testing.T) {
	type args struct {
		args []string
	}
	tests := []struct {
		name  string
		args  args
		want  []string
		want1 string
	}{
		{
			name: "empty",
			args: args{
				args: []string{},
			},
			want:  []string{},
			want1: "",
		},
		{
			name: "no fuzzy",
			args: args{
				args: []string{"test"},
			},
			want:  []string{"test"},
			want1: "",
		},
		{
			name: "single fuzzy",
			args: args{
				args: []string{"file:test"},
			},
			want:  []string{},
			want1: "test",
		},
		{
			name: "single fuzzy alternate",
			args: args{
				args: []string{"filename:test"},
			},
			want:  []string{},
			want1: "test",
		},
		{
			name: "multi fuzzy last wins",
			args: args{
				args: []string{"file:test", "file:other"},
			},
			want:  []string{},
			want1: "other",
		},
		{
			name: "single fuzzy single term",
			args: args{
				args: []string{"stuff", "file:test"},
			},
			want:  []string{"stuff"},
			want1: "test",
		},
		{
			name: "single fuzzy uppercase",
			args: args{
				args: []string{"FILE:test", "UPPER"},
			},
			want:  []string{"UPPER"},
			want1: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := PreParseQuery(tt.args.args)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PreParseQuery() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("PreParseQuery() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
