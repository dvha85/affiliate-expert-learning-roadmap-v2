//go:build !linux && !darwin

package main

import (
	"errors"
	"os"
)

func campaignOwnerCheck(os.FileInfo) error {
	return errors.New("campaign ownership validation unsupported on this OS")
}
