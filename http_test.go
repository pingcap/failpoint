// Copyright 2021 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package failpoint_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pingcap/failpoint"
	"github.com/stretchr/testify/require"
)

type badReader struct{}

func (badReader) Read([]byte) (int, error) {
	return 0, errors.New("mock bad read")
}

func TestServeHTTP(t *testing.T) {
	require.NoError(t, failpoint.Serve(":23389"))

	handler := &failpoint.HttpHandler{}

	// PUT
	req, err := http.NewRequest(http.MethodPut, "http://127.0.0.1/failpoint-name", strings.NewReader("return(1)"))
	require.NoError(t, err)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	require.Equal(t, http.StatusNoContent, res.Code)

	req, err = http.NewRequest(http.MethodPut, "http://127.0.0.1", strings.NewReader("return(1)"))
	require.NoError(t, err)
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	require.Equal(t, http.StatusBadRequest, res.Code)
	require.Contains(t, res.Body.String(), "malformed request URI")

	req, err = http.NewRequest(http.MethodPut, "http://127.0.0.1/failpoint-name", badReader{})
	require.NoError(t, err)
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	require.Equal(t, http.StatusBadRequest, res.Code)
	require.Contains(t, res.Body.String(), "failed ReadAll in PUT")

	req, err = http.NewRequest(http.MethodPut, "http://127.0.0.1/failpoint-name", strings.NewReader("invalid"))
	require.NoError(t, err)
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	require.Equal(t, http.StatusBadRequest, res.Code)
	require.Contains(t, res.Body.String(), "failed to set failpoint")

	// GET
	req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1/failpoint-name", strings.NewReader(""))
	require.NoError(t, err)
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	require.Equal(t, http.StatusOK, res.Code)
	require.Contains(t, res.Body.String(), "return(1)")

	req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1/failpoint-name-not-exists", strings.NewReader(""))
	require.NoError(t, err)
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	require.Equal(t, http.StatusNotFound, res.Code)
	require.Contains(t, res.Body.String(), "failed to GET")

	req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1/", strings.NewReader(""))
	require.NoError(t, err)
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	require.Equal(t, http.StatusOK, res.Code)
	require.Contains(t, res.Body.String(), "failpoint-name=return(1)")

	// DELETE
	req, err = http.NewRequest(http.MethodDelete, "http://127.0.0.1/failpoint-name", strings.NewReader(""))
	require.NoError(t, err)
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	require.Equal(t, http.StatusNoContent, res.Code)

	req, err = http.NewRequest(http.MethodDelete, "http://127.0.0.1/failpoint-name-not-exists", strings.NewReader(""))
	require.NoError(t, err)
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	require.Equal(t, http.StatusBadRequest, res.Code)
	require.Contains(t, res.Body.String(), "failed to delete failpoint")

	// DEFAULT
	req, err = http.NewRequest(http.MethodPost, "http://127.0.0.1/failpoint-name", strings.NewReader(""))
	require.NoError(t, err)
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	require.Equal(t, http.StatusMethodNotAllowed, res.Code)
	require.Contains(t, res.Body.String(), "Method not allowed")

	// Test environment variable injection
	resp, err := http.Get("http://127.0.0.1:23389/failpoint-env1")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), "return(10)")

	resp, err = http.Get("http://127.0.0.1:23389/failpoint-env2")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body, err = ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), "return(true)")
}
