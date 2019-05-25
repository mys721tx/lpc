package hosts_test

import (
	"testing"

	"github.com/mys721tx/lpc/pkg/hosts"
	"github.com/stretchr/testify/assert"
)

func TestParseLine(t *testing.T) {
	type args struct {
		line string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
		want2 string
	}{
		{
			name:  "standard",
			args:  args{line: "127.0.0.1 localhost"},
			want:  "127.0.0.1",
			want1: "localhost",
			want2: "",
		},
		{
			name:  "commentAllPound",
			args:  args{line: "#;234"},
			want:  "",
			want1: "",
			want2: ";234",
		},
		{
			name:  "commentAllSemicolon",
			args:  args{line: ";#234"},
			want:  "",
			want1: "",
			want2: "#234",
		},
		{
			name:  "commentPound",
			args:  args{line: "127.0.0.1 localhost #1234"},
			want:  "127.0.0.1",
			want1: "localhost",
			want2: "1234",
		},
		{
			name:  "commentSemicolon",
			args:  args{line: "127.0.0.1 localhost ;1234"},
			want:  "127.0.0.1",
			want1: "localhost",
			want2: "1234",
		},
		{
			name:  "IPv6",
			args:  args{line: "::1 localhost"},
			want:  "::1",
			want1: "localhost",
			want2: "",
		},
		{
			name:  "WhiteSpaceTab",
			args:  args{line: "127.0.0.1\tlocalhost\t"},
			want:  "127.0.0.1",
			want1: "localhost",
			want2: "",
		},
		{
			name:  "WhiteSpaceMixed",
			args:  args{line: "127.0.0.1 localhost\t"},
			want:  "127.0.0.1",
			want1: "localhost",
			want2: "",
		},
		{
			name:  "WhiteSpaceMultiple",
			args:  args{line: "127.0.0.1 \t  localhost\t  #123"},
			want:  "127.0.0.1",
			want1: "localhost",
			want2: "123",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := hosts.ParseLine(tt.args.line)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want1, got1)
			assert.Equal(t, tt.want2, got2)
		})
	}
}

func BenchmarkParseLine(b *testing.B) {
	for n := 0; n < b.N; n++ {
		hosts.ParseLine("127.0.0.1 localhost #test")
	}
}
