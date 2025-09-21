//go:build !unix

package main

func setUmask(mask int) {}
