//go:build unix

package main

import (
	"errors"
	"fmt"
	"os"
	"syscall"
)

// lockAll takes a non-blocking exclusive flock on every path (the same files tools/live.sh and the launcher
// hold while an instance is in use). It fails when one is held; release drops them all.
func lockAll(paths []string) (release func(), err error) {
	var held []*os.File
	release = func() {
		for _, f := range held {
			f.Close() // closing drops the flock
		}
	}
	for _, p := range paths {
		f, err := os.OpenFile(p, os.O_CREATE|os.O_RDWR, 0o666)
		if err != nil {
			release()
			return nil, err
		}
		if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
			f.Close()
			release()
			if errors.Is(err, syscall.EWOULDBLOCK) {
				return nil, fmt.Errorf("instance lock %s is held (a game, bot or tool run uses this prefix)", p)
			}
			return nil, err
		}
		held = append(held, f)
	}
	return release, nil
}
