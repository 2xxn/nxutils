package web

import (
	"reflect"
	"testing"
)

func TestRecognizePBFromHeaders(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string]string
		want    []string
	}{
		{
			name: "recognizes nginx and php from different headers",
			headers: map[string]string{
				"Server":       "nginx/1.24.0",
				"X-Powered-By": "PHP/8.0.30",
			},
			want: []string{PB_NGINX, PB_PHP},
		},
		{
			name: "recognizes old php versions",
			headers: map[string]string{
				"X-Powered-By": "PHP/5.6.40",
			},
			want: []string{PB_PHP_OLD},
		},
		{
			name: "recognizes iis from server or asp net",
			headers: map[string]string{
				"Server":       "Microsoft-IIS/10.0",
				"X-Powered-By": "ASP.NET",
			},
			want: []string{PB_IIS},
		},
		{
			name: "recognizes cloudflare",
			headers: map[string]string{
				"Server":          "cloudflare",
				"CF-RAY":          "123456",
				"CF-Cache-Status": "DYNAMIC",
			},
			want: []string{PB_CLOUDFLARE},
		},
		{
			name: "recognizes cloudfront",
			headers: map[string]string{
				"Via":         "1.1 abcdef.cloudfront.net (CloudFront)",
				"X-Amz-Cf-Id": "some-id",
			},
			want: []string{PB_CLOUDFRONT},
		},
		{
			name: "returns unknown when no known markers",
			headers: map[string]string{
				"Server": "custom-edge",
			},
			want: []string{PB_UNKNOWN},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RecognizePBFromHeaders(tt.headers)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("RecognizePBFromHeaders() = %v, want %v", got, tt.want)
			}
		})
	}
}
