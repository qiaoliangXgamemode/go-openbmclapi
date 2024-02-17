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
	"time"
)

type CacheOpt struct {
	Expiration time.Duration
}

type Cache interface {
	Set(key string, value string, opt CacheOpt)
	Get(key string) (value string, ok bool)
	SetBytes(key string, value []byte, opt CacheOpt)
	GetBytes(key string) (value []byte, ok bool)
}

type noCache struct{}

func (noCache) Set(key string, value string, opt CacheOpt)      {}
func (noCache) Get(key string) (value string, ok bool)          { return "", false }
func (noCache) SetBytes(key string, value []byte, opt CacheOpt) {}
func (noCache) GetBytes(key string) (value []byte, ok bool)     { return nil, false }

var NoCache Cache = (*noCache)(nil)
