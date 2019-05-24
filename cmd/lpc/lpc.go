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

	"github.com/miekg/dns"
)

var (
	pin, pout      string
	resolver, port string
)

const (
	prefix = "www."
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

	flag.StringVar(
		&resolver,
		"dns",
		"8.8.8.8",
		"IP address of the resolver, default to 8.8.8.8.",
	)

	flag.StringVar(
		&port,
		"port",
		"53",
		"port of the resolver, default to 53",
	)

	flag.Parse()

	addr := resolver + ":" + port

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
				if !strings.HasSuffix(f, ".") {
					f = f + "."
				}
				names[f] = true
			}
		}
	}

	if err := scn.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}

	c := new(dns.Client)

	for f, _ := range names {
		prefixed := prefix + f

		if _, exist := names[prefixed]; !exist && !strings.HasPrefix(f, prefix) {
			m := new(dns.Msg)
			m.SetQuestion(prefixed, dns.TypeA)
			in, _, err := c.Exchange(m, addr)

			if err != nil {
				fmt.Fprintln(
					os.Stderr,
					"error processing domain ",
					f,
					": ",
					err,
				)
			} else if len(in.Answer) > 0 {
				fmt.Fprintln(w, prefixed)
			}

		}
	}
}
