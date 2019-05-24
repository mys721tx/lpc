// lpc: Leaky Prefix Checker
// Copyright (C) 2019  Yishen Miao
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

var (
	pin, pout string
)

func main() {
	flag.StringVar(
		&pin,
		"in",
		"",
		"path to the host file, default to stdin.",
	)

	flag.StringVar(
		&pout,
		"out",
		"",
		"path to the output file, default to stdout.",
	)
	flag.Parse()

	var fin, fout *os.File

	if pin == "" {
		fin = os.Stdin
	} else if f, err := os.Open(pin); err == nil {
		fin = f
	} else {
		log.Panicf("failed to open %q: %v", pin, err)
	}

	if pout == "" {
		fout = os.Stdout
	} else if f, err := os.Create(pout); err == nil {
		fout = f
	} else {
		log.Panicf("failed to open %q: %v", pout, err)
	}

	defer func() {
		if err := fin.Close(); err != nil {
			log.Panicf("failed to close %q: %v", pin, err)
		}
	}()

	defer func() {
		if err := fout.Close(); err != nil {
			log.Panicf("failed to close %q: %v", pout, err)
		}
	}()

	scn := bufio.NewScanner(fin)
	w := bufio.NewWriter(fout)

	defer func() {
		if err := w.Flush(); err != nil {
			log.Panicf("failed to flush %q: %v", pout, err)
		}
	}()

	names := make(map[string]bool)

	for scn.Scan() {
		line := scn.Text()

		if line == "" ||
			strings.HasPrefix(line, "#") ||
			strings.HasPrefix(line, ";") {
			continue
		}

		for i, f := range strings.Fields(line) {
			if i != 0 &&
				!strings.HasPrefix(f, "#") &&
				!strings.HasPrefix(f, ";") {
				names[f] = true
			}
		}
	}

	if err := scn.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}

	for f, _ := range names {
		fmt.Fprintln(w, f)
	}
}
