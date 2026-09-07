//go:build linux || darwin

package main

import (
	"errors"
	"os"
	"syscall"
)

func campaignOwnerCheck(info os.FileInfo) error {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || stat.Uid != uint32(os.Geteuid()) {
		return errors.New("campaign directory owner mismatch")
	}
	return nil
}
