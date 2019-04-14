package failpoint_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/pingcap/check"
	"github.com/pingcap/failpoint"
)

func TestHttp(t *testing.T) {
	TestingT(t)
}

var _ = &httpSuite{}

type httpSuite struct{}

func (s httpSuite) TestServeHTTP(c *C) {
	handler := &failpoint.HttpHandler{}

	// PUT
	req, err := http.NewRequest(http.MethodPut, "http://127.0.0.1/failpoint-name", strings.NewReader("return(1)"))
	c.Assert(err, IsNil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	c.Assert(res.Code, Equals, http.StatusNoContent)

	req, err = http.NewRequest(http.MethodPut, "http://127.0.0.1/failpoint-name", strings.NewReader("invalid"))
	c.Assert(err, IsNil)
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	c.Assert(res.Code, Equals, http.StatusBadRequest)
	c.Assert(res.Body.String(), Matches, "failed to set failpoint")

	// GET
	req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1/failpoint-name", strings.NewReader(""))
	c.Assert(err, IsNil)
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	c.Assert(res.Code, Equals, http.StatusOK)
	c.Assert(res.Body.String(), Matches, `return\(1\)`)

	req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1/failpoint-name-not-exists", strings.NewReader(""))
	c.Assert(err, IsNil)
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	c.Assert(res.Code, Equals, http.StatusBadRequest)
	c.Assert(res.Body.String(), Matches, "failed to GET")

	req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1/", strings.NewReader(""))
	c.Assert(err, IsNil)
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	c.Assert(res.Code, Equals, http.StatusOK)
	c.Assert(res.Body.String(), Matches, `failpoint-name=return\(1\)`)

	// DELETE
	req, err = http.NewRequest(http.MethodDelete, "http://127.0.0.1/failpoint-name", strings.NewReader(""))
	c.Assert(err, IsNil)
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	c.Assert(res.Code, Equals, http.StatusNoContent)

	req, err = http.NewRequest(http.MethodDelete, "http://127.0.0.1/failpoint-name-not-exists", strings.NewReader(""))
	c.Assert(err, IsNil)
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	c.Assert(res.Code, Equals, http.StatusNoContent)
	c.Assert(res.Body.String(), Matches, "failed to delete failpoint")

	// DEFAULT
	req, err = http.NewRequest(http.MethodPost, "http://127.0.0.1/failpoint-name", strings.NewReader(""))
	c.Assert(err, IsNil)
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	c.Assert(res.Code, Equals, http.StatusMethodNotAllowed)
	c.Assert(res.Body.String(), Matches, "Method not allowed")
}
