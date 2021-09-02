/**
 * (C) Copyright IBM Corp. 2021
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package test

import (
	"flag"
	"fmt"
	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/ffyaml"
	"testing"
)

// FindConfigPath The path to the config may vary depending on which directory the tests are run from, so we have
// to figure out which one it's supposed to be.
func FindConfigPath(t *testing.T) string {
	possiblePaths := []string{
		"./src/common/config/goodConfig.yml",
		"../config/goodConfig.yml",
		"./goodConfig.yml",
		"common/config/goodConfig.yml",
	}
	fs := flag.NewFlagSet("servetest", flag.ContinueOnError)
	var args []string
	opts := []ff.Option{ff.WithConfigFileParser(ffyaml.Parser)}
	for _, path := range possiblePaths {
		opts = append(opts, ff.WithConfigFile(path))
		err := ff.Parse(fs, args, opts...)
		if err.Error() != fmt.Sprintf("open %s: no such file or directory", path) {
			return path
		}
	}
	t.Fatalf("COULD NOT FIND A CONFIG FILE!")
	return ""
}
