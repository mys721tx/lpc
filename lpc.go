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
	"net"
	"os"
	"strings"
	"time"

	"github.com/miekg/dns"

	"github.com/mys721tx/lpc/pkg/hosts"
)

var (
	pin, pout      string
	host, port     string
	tgt, prefix    string
	sleep, timeout int64
)

func bldJoin(b *strings.Builder, strs ...string) {

	for _, str := range strs {
		b.WriteString(str)
	}
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
		&host,
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
		&prefix,
		"prefix",
		"www.",
		"prefix to check for each hosts entry, default to www.",
	)

	flag.Int64Var(
		&timeout,
		"timeout",
		10,
		"timeout for each DNS query, default to 10s",
	)

	flag.StringVar(
		&tgt,
		"tgt",
		"0.0.0.0",
		"target IP address of the blocked entry, default to 0.0.0.0",
	)

	flag.Int64Var(
		&sleep,
		"sleep",
		100,
		"time between DNS query, default to 100ms",
	)

	flag.Parse()

	addr := net.JoinHostPort(host, port)

	var fin, fout *os.File

	if pin == "" {
		fin = os.Stdin
	} else if f, err := os.Open(pin); err != nil {
		log.Panicf("failed to open %q: %v", pin, err)
	} else {
		fin = f
	}

	if pout == "" {
		fout = os.Stdout
	} else if f, err := os.Create(pout); err != nil {
		log.Panicf("failed to open %q: %v", pout, err)
	} else {
		fout = f
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
	c.Timeout = time.Duration(timeout) * time.Second
	m := new(dns.Msg)

	for scn.Scan() {
		line := scn.Text()

		ip, hns, cmt := hosts.ParseLine(line)

		// Do not further process empty or commented line.
		if ip == "" {
			fmt.Fprintln(w, line)
			continue
		}

		// Process multi entry lines
		for _, fld := range hns {
			if _, exist := names[fld]; !exist {
				m.SetQuestion(dns.Fqdn(fld), dns.TypeA)
				in, _, err := c.Exchange(m, addr)
				time.Sleep(time.Duration(sleep) * time.Millisecond)

				var b strings.Builder

				bldJoin(&b, ip, " ", fld)

				if cmt != "" {
					bldJoin(&b, " #", cmt)
				}

				if err != nil {
					fmt.Fprintln(
						os.Stderr,
						"error processing domain",
						fld,
						err,
					)
				} else {

					if in.MsgHdr.Rcode != dns.RcodeSuccess {
						if cmt == "" {
							b.WriteString(" #")
						}
						b.WriteString(
							dns.RcodeToString[in.MsgHdr.Rcode],
						)
					}
				}

				if _, exist := names[fld]; !exist {
					fmt.Fprintln(w, b.String())
					names[fld] = true
				}
			}
		}

		for _, dom := range hns {
			domPfx := prefix + dom

			if _, exist := names[domPfx]; exist {
				continue
			}

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
