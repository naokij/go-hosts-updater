// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package main

func hostsPath() string {
	return "/etc/hosts"
}

func tmpPath() string {
	return "/tmp"
}

func newline() string {
	return "\n"
}
