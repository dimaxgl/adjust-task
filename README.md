# Adjust Take Home Challenge

Tool which makes http requests and prints the address of the request along with the MD5 hash of the response.

## Requirements

- go 1.17

## Build

```bash
go build -o task cmd/main.go
```

## Command-line Options

Supported command-line options:

```bash
./task -h
Usage: ./task [OPTIONS] url1 url2 ...
  -parallel int
        parallel processing value (default 10)
```