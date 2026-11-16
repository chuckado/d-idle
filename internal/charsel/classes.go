// Package charsel is the backend of the Character Select screen (docs/charselect.md): it lists the characters of a save
// folder from the .d2s header fields only, validates names, archives before it deletes, restores, parks the per
// character bot state on a switch, and carries the Manager the /api/characters routes call.
package charsel

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed classes.json
var classesJSON []byte

// Class is one playable class. The numbers live in classes.json (build flexibility: no class literal in Go).
type Class struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	D2S      int    `json:"d2s"`
	D2DIndex int    `json:"d2d_index"`
}

var classes = func() []Class {
	var f struct {
		Classes []Class `json:"classes"`
	}
	if err := json.Unmarshal(classesJSON, &f); err != nil || len(f.Classes) == 0 {
		panic("charsel: classes.json: " + fmt.Sprint(err))
	}
	return f.Classes
}()

// Classes returns the class table.
func Classes() []Class { return append([]Class(nil), classes...) }

// ClassByID finds a class by id or display name, case-insensitively. An unknown class is an error naming it; there is
// no fallback class.
func ClassByID(id string) (Class, error) {
	for _, c := range classes {
		if strings.EqualFold(c.ID, id) || strings.EqualFold(c.Name, id) {
			return c, nil
		}
	}
	return Class{}, fmt.Errorf("charsel: unknown class %q", id)
}

// ClassByD2S finds a class by the class byte of a save header.
func ClassByD2S(b int) (Class, error) {
	for _, c := range classes {
		if c.D2S == b {
			return c, nil
		}
	}
	return Class{}, fmt.Errorf("charsel: unknown class byte %d in the save header", b)
}
