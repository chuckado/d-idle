//go:build !unix

package main

import "errors"

func lockAll([]string) (func(), error) {
	return nil, errors.New("d2restore: instance locks need a unix host")
}
