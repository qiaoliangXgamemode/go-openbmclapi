/**
 * OpenBmclAPI (Golang Edition)
 * Copyright (C) 2024 Kevin Z <zyxkad@gmail.com>
 * All rights reserved
 *
 *  This program is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU Affero General Public License as published
 *  by the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU Affero General Public License for more details.
 *
 *  You should have received a copy of the GNU Affero General Public License
 *  along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package main

import (
	"io"
	"sync/atomic"
	"time"
)

type LimitedReadWriteCloser struct {
	io.ReadWriteCloser
	controller *RateController

	closed               atomic.Bool
	ReadWriteCloserAfter time.Time
}

var _ io.ReadCloser = (*LimitedReader)(nil)

func (r *LimitedReader) Read(buf []byte) (n int, err error) {
	if !r.readAfter.IsZero() {
		now := time.Now()
		if dur := r.readAfter.Sub(now); dur > 0 {
			time.Sleep(dur)
		}
	}
	m := r.controller.preRead(len(buf))
	n, err = r.Reader.Read(buf[:m])
	dur := r.controller.afterRead(n, m-n)
	if dur > 0 {
		r.readAfter = time.Now().Add(dur)
	} else {
		r.readAfter = time.Time{}
	}
	return
}
