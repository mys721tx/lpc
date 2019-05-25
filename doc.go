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

/*
	LPC checks a hosts block list and add of a given prefix.

	The Leaky Prefix Checker for domain names that escapes a hosts block
	list through `www.` prefix.

	Usage:
		lpc [flags]
	The flags are:
		-dns string
			IP address of the resolver, default to 8.8.8.8.",
		-in string
			path to the hosts file, default to stdin.
		-port string
			port of the resolver, default to 53
		-out string
			path to the output file, default to stdout.
		-tgt string
			target IP address of the blocked entry, default to 0.0.0.0
	Example:
		cat /etc/hosts | lpc
		lpc -in /etc/hosts -out hosts.tmp
*/
package main
