/*
 * Copyright (C) 2018 Onchain <onchain@onchain.com>
 *
 * This file is part of The ontology_Zero.
 *
 * The ontology_Zero is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology_Zero is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology_Zero.  If not, see <http://www.gnu.org/licenses/>.
 */

package trie

func hexToCompact(hex []byte) []byte {
	terminator := byte(0)
	if hasTerm(hex) {
		terminator = 1
		hex = hex[:len(hex) - 1]
	}
	buf := make([]byte, len(hex) / 2 + 1)
	buf[0] = terminator << 5
	if len(hex) & 1 == 1 {
		buf[0] |= 1 << 4
		buf[0] |= hex[0]
		hex = hex[1:]
	}
	decodeNibbles(hex, buf[1:])
	return buf
}

func decodeNibbles(nibbles []byte, bytes []byte) {
	for bi, ni := 0, 0; ni < len(nibbles); bi, ni = bi + 1, ni + 2 {
		bytes[bi] = nibbles[ni] << 4 | nibbles[ni + 1]
	}
}

func compactToHex(compact []byte) []byte {
	base := keyBytesToHex(compact)
	base = base[:len(base) - 1]
	if base[0] >= 2 {
		base = append(base, 16)
	}
	chop := 2 - base[0] & 1
	return base[chop:]
}

func keyBytesToHex(str []byte) []byte {
	l := len(str) * 2 + 1
	var nibbles = make([]byte, l)
	for i, b := range str {
		nibbles[i * 2] = b / 16
		nibbles[i * 2 + 1] = b % 16
	}
	nibbles[l - 1] = 16
	return nibbles
}

func prefixLen(a, b []byte) int {
	var i, length = 0, len(a)
	if len(b) < length {
		length = len(b)
	}
	for ; i < length; i++ {
		if a[i] != b[i] {
			break
		}
	}
	return i
}

func hasTerm(s []byte) bool {
	return len(s) > 0 && s[len(s) - 1] == 16
}