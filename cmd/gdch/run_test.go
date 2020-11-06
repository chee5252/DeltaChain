// Copyright 2016 The go-deltachaineum Authors
// This file is part of go-deltachaineum.
//
// go-deltachaineum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-deltachaineum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-deltachaineum. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/docker/docker/pkg/reexec"
	"github.com/deltachaineum/go-deltachaineum/internal/cmdtest"
)

func tmpdir(t *testing.T) string {
	dir, err := ioutil.TempDir("", "gdch-test")
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

type testgdch struct {
	*cmdtest.TestCmd

	// template variables for expect
	Datadir   string
	Deltachainbase string
}

func init() {
	// Run the app if we've been exec'd as "gdch-test" in runGdch.
	reexec.Register("gdch-test", func() {
		if err := app.Run(os.Args); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	})
}

func TestMain(m *testing.M) {
	// check if we have been reexec'd
	if reexec.Init() {
		return
	}
	os.Exit(m.Run())
}

// spawns gdch with the given command line args. If the args don't set --datadir, the
// child g gets a temporary data directory.
func runGdch(t *testing.T, args ...string) *testgdch {
	tt := &testgdch{}
	tt.TestCmd = cmdtest.NewTestCmd(t, tt)
	for i, arg := range args {
		switch {
		case arg == "-datadir" || arg == "--datadir":
			if i < len(args)-1 {
				tt.Datadir = args[i+1]
			}
		case arg == "-deltachainbase" || arg == "--deltachainbase":
			if i < len(args)-1 {
				tt.Deltachainbase = args[i+1]
			}
		}
	}
	if tt.Datadir == "" {
		tt.Datadir = tmpdir(t)
		tt.Cleanup = func() { os.RemoveAll(tt.Datadir) }
		args = append([]string{"-datadir", tt.Datadir}, args...)
		// Remove the temporary datadir if somdching fails below.
		defer func() {
			if t.Failed() {
				tt.Cleanup()
			}
		}()
	}

	// Boot "gdch". This actually runs the test binary but the TestMain
	// function will prevent any tests from running.
	tt.Run("gdch-test", args...)

	return tt
}
