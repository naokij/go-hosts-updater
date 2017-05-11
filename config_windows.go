package main

import "os"

func hostsPath() string {
	return `C:\Windows\System32\drivers\etc\hosts`
}

func tmpPath() string {
	return os.Getenv("temp")
}
