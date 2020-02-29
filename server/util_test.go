package server

import "testing"

func TestFindHost1(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name       string
		args       args
		wantMethod string
		wantHost   string
		wantErr    bool
	}{
		{
			name:       "find host1",
			args:       args{data: []byte("GET http://www.baidu.com/ HTTP/1.1\r\nHost: www.baidu.com\r\n")},
			wantMethod: "GET",
			wantHost:   "www.baidu.com:80",
			wantErr:    false,
		},
		{
			name:       "find host2",
			args:       args{data: []byte("GET / HTTP/1.1\r\nHost: www.baidu.com\r\n")},
			wantMethod: "GET",
			wantHost:   "www.baidu.com:80",
			wantErr:    false,
		},
		{
			name:       "find host2",
			args:       args{data: []byte("GET / HTTP/1.1\r\nHost: www.baidu.com:3333\r\n")},
			wantMethod: "GET",
			wantHost:   "www.baidu.com:3333",
			wantErr:    false,
		},
		{
			name:       "find host3",
			args:       args{data: []byte("CONNECT / HTTP/1.1\r\nHost: www.baidu.com\r\n")},
			wantMethod: "CONNECT",
			wantHost:   "www.baidu.com:443",
			wantErr:    false,
		},
		{
			name:       "find host3",
			args:       args{data: []byte("CONNECT / HTTP/1.1\r\nHost: www.baidu.com:3333\r\n")},
			wantMethod: "CONNECT",
			wantHost:   "www.baidu.com:3333",
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMethod, gotHost, err := FindHost(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindHost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotMethod != tt.wantMethod {
				t.Errorf("FindHost() gotMethod = %v, want %v", gotMethod, tt.wantMethod)
			}
			if gotHost != tt.wantHost {
				t.Errorf("FindHost() gotHost = %v, want %v", gotHost, tt.wantHost)
			}
		})
	}
}
