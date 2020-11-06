// Copyright 2017 The go-deltachaineum Authors
// This file is part of the go-deltachaineum library.
//
// The go-deltachaineum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-deltachaineum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-deltachaineum library. If not, see <http://www.gnu.org/licenses/>.

package rpc

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPErrorResponseWithDelete(t *testing.T) {
	testHTTPErrorResponse(t, http.MdchodDelete, contentType, "", http.StatusMdchodNotAllowed)
}

func TestHTTPErrorResponseWithPut(t *testing.T) {
	testHTTPErrorResponse(t, http.MdchodPut, contentType, "", http.StatusMdchodNotAllowed)
}

func TestHTTPErrorResponseWithMaxContentLength(t *testing.T) {
	body := make([]rune, maxRequestContentLength+1)
	testHTTPErrorResponse(t,
		http.MdchodPost, contentType, string(body), http.StatusRequestEntityTooLarge)
}

func TestHTTPErrorResponseWithEmptyContentType(t *testing.T) {
	testHTTPErrorResponse(t, http.MdchodPost, "", "", http.StatusUnsupportedMediaType)
}

func TestHTTPErrorResponseWithValidRequest(t *testing.T) {
	testHTTPErrorResponse(t, http.MdchodPost, contentType, "", 0)
}

func testHTTPErrorResponse(t *testing.T, mdchod, contentType, body string, expected int) {
	request := httptest.NewRequest(mdchod, "http://url.com", strings.NewReader(body))
	request.Header.Set("content-type", contentType)
	if code, _ := validateRequest(request); code != expected {
		t.Fatalf("response code should be %d not %d", expected, code)
	}
}
