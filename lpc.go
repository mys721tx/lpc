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
	"time"

	"github.com/miekg/dns"
)

var (
	pin, pout      string
	resolver, port string
	tgt            string
	sleep          int64
)

const (
	prefix = "www."
)

// splitLine takes a hosts file line and returns it in two parts. The second
// part is the comment, started with "#" or ";", which ever comes first.
func splitLine(line string) (string, string) {
	iPnd := strings.Index(line, "#")
	iSmc := strings.Index(line, ";")

	if iPnd < iSmc {
		if iPnd > -1 {
			return line[:iPnd], line[iPnd:]
		} else {
			return line[:iSmc], line[iSmc:]
		}
	} else if iSmc < iPnd {
		if iSmc > -1 {
			return line[:iSmc], line[iSmc:]
		} else {
			return line[:iPnd], line[iPnd:]
		}
	}
	return line, ""
}

func main() {
	flag.StringVar(
		&pin,
		"in",
		"",
		"path to the hosts file, default to stdin.",
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

	flag.StringVar(
		&tgt,
		"tgt",
		"0.0.0.0",
		"target IP address of the blocked entry, default to 0.0.0.0",
	)

	flag.Int64Var(
		&sleep,
		"time",
		100,
		"time between DNS query, default to 100ms",
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

	c := new(dns.Client)

	for scn.Scan() {
		tmpNames := make(map[string]bool)

		line := scn.Text()

		// Do not further process empty or commented line.
		if line == "" ||
			strings.HasPrefix(line, "#") ||
			strings.HasPrefix(line, ";") {
			fmt.Fprintln(w, line)
			continue
		}

		entry, comment := splitLine(line)

		// Process multi entry lines
		for i, fld := range strings.Fields(entry) {
			if i != 0 {
				if written, exist := names[fld]; !exist || !written {
					m := new(dns.Msg)
					m.SetQuestion(dns.Fqdn(fld), dns.TypeA)
					in, _, err := c.Exchange(m, addr)
					time.Sleep(time.Duration(sleep) * time.Millisecond)

					if err != nil {
						fmt.Fprintln(
							os.Stderr,
							"error processing domain",
							fld,
							err,
						)
						names[fld] = false
					} else {
						var b strings.Builder

						b.WriteString(
							fmt.Sprintf(
								"%s %s",
								tgt,
								strings.TrimSuffix(fld, "."),
							),
						)

						if comment != "" {
							b.WriteString(" ")
							b.WriteString(comment)
						}

						if in.MsgHdr.Rcode != dns.RcodeSuccess {
							if comment == "" {
								b.WriteString(" #")
							}
							b.WriteString(
								dns.RcodeToString[in.MsgHdr.Rcode],
							)
						}
						fmt.Fprintln(w, b.String())
						names[fld] = true
					}
				}

				if !strings.HasPrefix(fld, prefix) {
					tmpNames[fld] = true
				}
			}
		}

		for dom, _ := range tmpNames {
			domPfx := prefix + dom

			if _, exist := names[domPfx]; exist {
				continue
			}

			names[domPfx] = false

			m := new(dns.Msg)
			m.SetQuestion(dns.Fqdn(domPfx), dns.TypeA)
			in, _, err := c.Exchange(m, addr)
			time.Sleep(time.Duration(sleep) * time.Millisecond)

			if err != nil {
				fmt.Fprintln(
					os.Stderr,
					"error processing domain",
					domPfx,
					err,
				)
			} else if in.MsgHdr.Rcode == dns.RcodeSuccess {
				fmt.Fprintln(w, tgt, domPfx)
				names[domPfx] = true
			} else {
				fmt.Fprintln(
					os.Stderr,
					"error processing domain",
					domPfx,
					dns.RcodeToString[in.MsgHdr.Rcode],
				)
			}

		}
	}

	if err := scn.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}
