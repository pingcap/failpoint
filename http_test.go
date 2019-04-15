package failpoint_test

import (
	"errors"
	"fmt"
	"io/ioutil"
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

var _ = Suite(&httpSuite{})

type httpSuite struct{}

type hasPrefix struct {
	*CheckerInfo
}

var Contains Checker = &hasPrefix{
	&CheckerInfo{Name: "Contains", Params: []string{"obtained", "expected"}},
}

func (checker *hasPrefix) Check(params []interface{}, names []string) (result bool, error string) {
	defer func() {
		if v := recover(); v != nil {
			result = false
			error = fmt.Sprint(v)
		}
	}()
	return strings.Contains(params[0].(string), params[1].(string)), ""
}

type badReader struct{}

func (badReader) Read([]byte) (int, error) {
	return 0, errors.New("mock bad read")
}

func (s *httpSuite) TestServeHTTP(c *C) {
	handler := &failpoint.HttpHandler{}

	// PUT
	req, err := http.NewRequest(http.MethodPut, "http://127.0.0.1/failpoint-name", strings.NewReader("return(1)"))
	c.Assert(err, IsNil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	c.Assert(res.Code, Equals, http.StatusNoContent)

	req, err = http.NewRequest(http.MethodPut, "http://127.0.0.1", strings.NewReader("return(1)"))
	c.Assert(err, IsNil)
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	c.Assert(res.Code, Equals, http.StatusBadRequest)
	c.Assert(res.Body.String(), Contains, "malformed request URI")

	req, err = http.NewRequest(http.MethodPut, "http://127.0.0.1/failpoint-name", badReader{})
	c.Assert(err, IsNil)
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	c.Assert(res.Code, Equals, http.StatusBadRequest)
	c.Assert(res.Body.String(), Contains, "failed ReadAll in PUT")

	req, err = http.NewRequest(http.MethodPut, "http://127.0.0.1/failpoint-name", strings.NewReader("invalid"))
	c.Assert(err, IsNil)
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	c.Assert(res.Code, Equals, http.StatusBadRequest)
	c.Assert(res.Body.String(), Contains, "failed to set failpoint")

	// GET
	req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1/failpoint-name", strings.NewReader(""))
	c.Assert(err, IsNil)
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	c.Assert(res.Code, Equals, http.StatusOK)
	c.Assert(res.Body.String(), Contains, "return(1)")

	req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1/failpoint-name-not-exists", strings.NewReader(""))
	c.Assert(err, IsNil)
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	c.Assert(res.Code, Equals, http.StatusNotFound)
	c.Assert(res.Body.String(), Contains, "failed to GET")

	req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1/", strings.NewReader(""))
	c.Assert(err, IsNil)
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	c.Assert(res.Code, Equals, http.StatusOK)
	c.Assert(res.Body.String(), Contains, "failpoint-name=return(1)")

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
	c.Assert(res.Code, Equals, http.StatusBadRequest)
	c.Assert(res.Body.String(), Contains, "failed to delete failpoint")

	// DEFAULT
	req, err = http.NewRequest(http.MethodPost, "http://127.0.0.1/failpoint-name", strings.NewReader(""))
	c.Assert(err, IsNil)
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	c.Assert(res.Code, Equals, http.StatusMethodNotAllowed)
	c.Assert(res.Body.String(), Contains, "Method not allowed")

	// Test environment variable injection
	resp, err := http.Get("http://127.0.0.1:23389/failpoint-env")
	c.Assert(err, IsNil)
	c.Assert(resp.StatusCode, Equals, http.StatusOK)
	body, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, IsNil)
	c.Assert(string(body), Contains, "return(10)")
}
