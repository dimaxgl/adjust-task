package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
)

var ts *httptest.Server

func TestMain(t *testing.M) {
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/hash1" {
			_, _ = w.Write([]byte("hash1"))
		}
		if r.RequestURI == "/hash2" {
			_, _ = w.Write([]byte("hash2"))
		}
	}))
	os.Exit(t.Run())
}

func Test_prepareUrl(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Test without scheme", args{url: "google.com"}, "http://google.com"},
		{"Test with http scheme", args{url: "http://icloud.com"}, "http://icloud.com"},
		{"Test with https scheme", args{url: "https://github.com"}, "https://github.com"},
		{"Test with https scheme and query", args{url: "http://github.com/dimaxgl"}, "http://github.com/dimaxgl"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := prepareUrl(tt.args.url); got != tt.want {
				t.Errorf("prepareUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_processUrl(t *testing.T) {
	type args struct {
		in  string
		out [2]string
	}
	tests := []struct {
		name string
		args args
	}{
		{"", args{in: ts.URL + "/hash1", out: [2]string{ts.URL + "/hash1", "00c6ee2e21a7548de6260cf72c4f4b5b"}}},
		{"", args{in: ts.URL + "/hash2", out: [2]string{ts.URL + "/hash2", "58833651db311ba4bc11cb26b1900b0f"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				wg      sync.WaitGroup
				inChan  = make(chan string)
				outChan = make(chan [2]string)
			)
			wg.Add(1)
			go processUrl(&wg, inChan, outChan)
			inChan <- tt.args.in
			close(inChan)
			if got := <-outChan; got != tt.args.out {
				t.Errorf("getHashedUrlContent() got = %v, want %v", got, tt.args.out)
			}
		})
	}
}

func Test_getHashedUrlContent(t *testing.T) {
	type args struct {
		cli *http.Client
		url string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"Test hash1", args{url: ts.URL + "/hash1", cli: ts.Client()}, "00c6ee2e21a7548de6260cf72c4f4b5b", false},
		{"Test hash2", args{url: ts.URL + "/hash2", cli: ts.Client()}, "58833651db311ba4bc11cb26b1900b0f", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getHashedUrlContent(tt.args.cli, tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("getHashedUrlContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getHashedUrlContent() for url = %s got = %v, want %v", tt.args.url, got, tt.want)
			}
		})
	}
}

func TestMainFunc(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	os.Args = []string{"-parallel", "10", ts.URL + "/hash1", ts.URL + "/hash2"}
	main()
	_ = w.Close()
	out, _ := ioutil.ReadAll(r)
	if !bytes.Contains(out, []byte(ts.URL+"/hash1"+" "+"00c6ee2e21a7548de6260cf72c4f4b5b")) {
		t.Errorf("TestMainFunc() hash1 not found")
	}
	if !bytes.Contains(out, []byte(ts.URL+"/hash2"+" "+"58833651db311ba4bc11cb26b1900b0f")) {
		t.Errorf("TestMainFunc() hash2 not found")
	}
	os.Stdout = oldStdout
}
