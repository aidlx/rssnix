//go:build unix

package main

import "golang.org/x/sys/unix"

func setUmask(mask int) {
	unix.Umask(mask)
}
