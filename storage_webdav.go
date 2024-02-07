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
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"

	"github.com/emersion/go-webdav"
	"gopkg.in/yaml.v3"

	"github.com/LiterMC/go-openbmclapi/internal/gosrc"
)

type basicWebDavStorageOption struct {
}

type WebDavStorageOption struct {
	PreGenMeasures bool `yaml:"pre-gen-measures"`

	Alias     string      `yaml:"alias,omitempty"`
	aliasUser *WebDavUser `yaml:"-"`

	WebDavUser `yaml:",inline,omitempty"`
}

var (
	_ yaml.Marshaler   = (*WebDavStorageOption)(nil)
	_ yaml.Unmarshaler = (*WebDavStorageOption)(nil)
)

func (o *WebDavStorageOption) MarshalYAML() (any, error) {
	type T WebDavStorageOption
	return (*T)(o), nil
}

func (o *WebDavStorageOption) UnmarshalYAML(n *yaml.Node) (err error) {
	o.Alias = ""
	type T WebDavStorageOption
	if err = n.Decode((*T)(o)); err != nil {
		return
	}
	return
}

func (o *WebDavStorageOption) GetEndPoint() string {
	if o.EndPoint != "" {
		return o.EndPoint
	}
	if o.Alias != "" {
		// assert o.aliasUser != nil
		return o.aliasUser.EndPoint
	}
	return ""
}

func (o *WebDavStorageOption) GetUsername() string {
	if o.Username != "" {
		return o.Username
	}
	if o.Alias != "" {
		// assert o.aliasUser != nil
		return o.aliasUser.Username
	}
	return ""
}

func (o *WebDavStorageOption) GetPassword() string {
	if o.Password != "" {
		return o.Password
	}
	if o.Alias != "" {
		// assert o.aliasUser != nil
		return o.aliasUser.Password
	}
	return ""
}

type WebDavStorage struct {
	opt WebDavStorageOption

	cli *webdav.Client
}

var _ Storage = (*WebDavStorage)(nil)

func init() {
	RegisterStorageFactory(StorageWebdav, StorageFactory{
		New:       func() Storage { return new(WebDavStorage) },
		NewConfig: func() any { return new(WebDavStorageOption) },
	})
}

func (s *WebDavStorage) String() string {
	return fmt.Sprintf("<WebDavStorage endpoint=%q user=%s>", s.opt.GetEndPoint(), s.opt.GetUsername())
}

func (s *WebDavStorage) Options() any {
	return &s.opt
}

func (s *WebDavStorage) SetOptions(newOpts any) {
	s.opt = *(newOpts.(*WebDavStorageOption))
}

func (s *WebDavStorage) Init(ctx context.Context) (err error) {
	if alias := s.opt.Alias; alias != "" {
		user, ok := config.WebdavUsers[alias]
		if !ok {
			logErrorf("Web dav user %q does not exists", alias)
			os.Exit(1)
		}
		s.opt.aliasUser = user
	}

	if s.cli, err = webdav.NewClient(
		webdav.HTTPClientWithBasicAuth(http.DefaultClient, s.opt.GetUsername(), s.opt.GetPassword()),
		s.opt.GetEndPoint()); err != nil {
		return
	}

	if err = s.cli.Mkdir(ctx, "measure"); err != nil {
		logErrorf("Could not create measure folder %v", err)
	}
	if s.opt.PreGenMeasures {
		logInfo("Creating measure files")
		for i := 1; i <= 200; i++ {
			if err := s.createMeasureFile(ctx, i); err != nil {
				os.Exit(2)
			}
		}
		logInfo("Measure files created")
	}
	return
}

func (s *WebDavStorage) hashToPath(hash string) string {
	return path.Join(hash[0:2], hash)
}

func (s *WebDavStorage) Size(hash string) (int64, error) {
	stat, err := os.Stat(s.hashToPath(hash))
	if err != nil {
		return 0, err
	}
	return stat.Size(), nil
}

func (s *WebDavStorage) Open(hash string) (io.ReadCloser, error) {
	return s.cli.Open(context.Background(), s.hashToPath(hash))
}

func (s *WebDavStorage) Create(hash string) (io.WriteCloser, error) {
	return s.cli.Create(context.Background(), s.hashToPath(hash))
}

func (s *WebDavStorage) Remove(hash string) error {
	return s.cli.RemoveAll(context.Background(), s.hashToPath(hash))
}

func (s *WebDavStorage) WalkDir(walker func(hash string) error) error {
	return nil
}

var noRedirectCli = &http.Client{
	CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func (s *WebDavStorage) serveWithRedirectIfPossible(rw http.ResponseWriter, req *http.Request, size int64, path string) (int64, error) {
	target, err := url.JoinPath(s.opt.GetEndPoint(), path)
	if err != nil {
		return 0, err
	}
	tgReq, err := http.NewRequestWithContext(req.Context(), http.MethodHead, target, nil)
	if err != nil {
		return 0, err
	}
	rangeH := req.Header.Get("Range")
	if rangeH != "" {
		tgReq.Header.Set("Range", rangeH)
	}
	copyHeader("If-Modified-Since", tgReq.Header, req.Header)
	copyHeader("If-Unmodified-Since", tgReq.Header, req.Header)
	copyHeader("If-None-Match", tgReq.Header, req.Header)
	copyHeader("If-Match", tgReq.Header, req.Header)
	copyHeader("If-Range", tgReq.Header, req.Header)
	resp, err := noRedirectCli.Do(tgReq)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	rwh := rw.Header()
	switch resp.StatusCode / 100 {
	case 3:
		// fix the size for Ranged request
		rgs, err := gosrc.ParseRange(rangeH, size)
		if err == nil && len(rgs) > 0 {
			var newSize int64 = 0
			for _, r := range rgs {
				newSize += r.Length
			}
			if newSize < size {
				size = newSize
			}
		}
		copyHeader("Location", rwh, resp.Header)
		copyHeader("ETag", rwh, resp.Header)
		copyHeader("Last-Modified", rwh, resp.Header)
		rw.WriteHeader(resp.StatusCode)
		return size, nil
	case 2:
		copyHeader("ETag", rwh, resp.Header)
		copyHeader("Last-Modified", rwh, resp.Header)
		copyHeader("Content-Length", rwh, resp.Header)
		copyHeader("Content-Range", rwh, resp.Header)
		rw.WriteHeader(resp.StatusCode)
		n, _ := io.Copy(rw, resp.Body)
		return n, nil
	default:
		return 0, webdav.NewHTTPError(resp.StatusCode, nil)
	}
}

func copyHeader(key string, dst, src http.Header) {
	v := src.Get(key)
	if v != "" {
		dst.Set(key, v)
	}
}

func (s *WebDavStorage) ServeDownload(rw http.ResponseWriter, req *http.Request, hash string, size int64) (int64, error) {
	target, err := url.JoinPath(s.opt.GetEndPoint(), "download", hash[0:2], hash)
	if err != nil {
		return 0, err
	}
	tgReq, err := http.NewRequestWithContext(req.Context(), http.MethodGet, target, nil)
	if err != nil {
		return 0, err
	}
	rangeH := req.Header.Get("Range")
	if rangeH != "" {
		tgReq.Header.Set("Range", rangeH)
	}
	copyHeader("If-Modified-Since", tgReq.Header, req.Header)
	copyHeader("If-Unmodified-Since", tgReq.Header, req.Header)
	copyHeader("If-None-Match", tgReq.Header, req.Header)
	copyHeader("If-Match", tgReq.Header, req.Header)
	copyHeader("If-Range", tgReq.Header, req.Header)
	resp, err := noRedirectCli.Do(tgReq)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	rwh := rw.Header()
	switch resp.StatusCode / 100 {
	case 3:
		// fix the size for Ranged request
		rgs, err := gosrc.ParseRange(rangeH, size)
		if err == nil && len(rgs) > 0 {
			var newSize int64 = 0
			for _, r := range rgs {
				newSize += r.Length
			}
			if newSize < size {
				size = newSize
			}
		}
		copyHeader("Location", rwh, resp.Header)
		copyHeader("ETag", rwh, resp.Header)
		copyHeader("Last-Modified", rwh, resp.Header)
		rw.WriteHeader(resp.StatusCode)
		return size, nil
	case 2:
		copyHeader("ETag", rwh, resp.Header)
		copyHeader("Last-Modified", rwh, resp.Header)
		copyHeader("Content-Length", rwh, resp.Header)
		copyHeader("Content-Range", rwh, resp.Header)
		rw.WriteHeader(resp.StatusCode)
		n, _ := io.Copy(rw, resp.Body)
		return n, nil
	default:
		return 0, webdav.NewHTTPError(resp.StatusCode, nil)
	}
}

func (s *WebDavStorage) ServeMeasure(rw http.ResponseWriter, req *http.Request, size int) error {
	if err := s.createMeasureFile(req.Context(), size); err != nil {
		return err
	}
	target, err := url.JoinPath(s.opt.GetEndPoint(), "measure", strconv.Itoa(size))
	if err != nil {
		return err
	}
	tgReq, err := http.NewRequestWithContext(req.Context(), http.MethodHead, target, nil)
	if err != nil {
		return err
	}
	rangeH := req.Header.Get("Range")
	if rangeH != "" {
		tgReq.Header.Set("Range", rangeH)
	}
	copyHeader("If-Modified-Since", tgReq.Header, req.Header)
	copyHeader("If-Unmodified-Since", tgReq.Header, req.Header)
	copyHeader("If-None-Match", tgReq.Header, req.Header)
	copyHeader("If-Match", tgReq.Header, req.Header)
	copyHeader("If-Range", tgReq.Header, req.Header)
	resp, err := noRedirectCli.Do(tgReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	rwh := rw.Header()
	switch resp.StatusCode / 100 {
	case 3:
		copyHeader("Location", rwh, resp.Header)
		copyHeader("ETag", rwh, resp.Header)
		copyHeader("Last-Modified", rwh, resp.Header)
		rw.WriteHeader(resp.StatusCode)
		return nil
	case 2:
		fallthrough
	default:
		rw.Header().Set("Content-Length", strconv.Itoa(size*mbChunkSize))
		rw.WriteHeader(http.StatusOK)
		if req.Method == http.MethodGet {
			for i := 0; i < size; i++ {
				rw.Write(mbChunk[:])
			}
		}
		return nil
	}
}

func (s *WebDavStorage) createMeasureFile(ctx context.Context, size int) (err error) {
	t := path.Join("measure", strconv.Itoa(size))
	if stat, err := s.cli.Stat(ctx, t); err == nil {
		tsz := (int64)(size) * mbChunkSize
		if size == 0 {
			tsz = 2
		}
		if stat.Size == tsz {
			return nil
		}
		logDebugf("File [%d] size %d does not match %d", size, stat.Size, tsz)
	} else if ctx.Err() != nil {
		return ctx.Err()
	} else if !errors.Is(err, os.ErrNotExist) {
		logErrorf("Cannot get stat of %s: %v", t, err)
	}
	logInfof("Creating measure file at %q", t)
	w, err := s.cli.Create(ctx, t)
	if err != nil {
		logErrorf("Cannot create measure file %q: %v", t, err)
		return
	}
	defer w.Close()
	if size == 0 {
		if _, err = w.Write(mbChunk[:2]); err != nil {
			logErrorf("Cannot write measure file %q: %v", t, err)
			return
		}
	} else {
		for j := 0; j < size; j++ {
			if _, err = w.Write(mbChunk[:]); err != nil {
				logErrorf("Cannot write measure file %q: %v", t, err)
				return
			}
		}
	}
	return nil
}