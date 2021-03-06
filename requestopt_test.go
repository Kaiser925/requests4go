// Developed by Kaiser925 on 2021/1/25.
// Lasted modified 2021/1/25.
// Copyright (c) 2021.  All rights reserved
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package requests4go

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestParams(t *testing.T) {
	params := map[string]string{
		"b": "b",
		"a": "c",
	}
	req, _ := NewRequest("get", "http://simple.org/path/?b=a", Params(params))
	assert.Equal(t, req.URL.String(), "http://simple.org/path/?a=c&b=b")
}

var authTests = []struct {
	username, password string
	ok                 bool
}{
	{"Aladdin", "open sesame", true},
	{"Aladdin", "open:sesame", true},
	{"", "", true},
}

func TestAuth(t *testing.T) {
	for _, tt := range authTests {
		r, _ := NewRequest("GET", "http://example.com/", Auth(tt.username, tt.password))
		username, password, ok := r.BasicAuth()
		assert.Equal(t, ok, tt.ok)
		assert.Equal(t, username, tt.username)
		assert.Equal(t, password, tt.password)
	}
}

var headerTests = []struct {
	k string
	v string
}{
	{"apple", "ok"},
	{"banana", "okokok"},
}

func TestHeaders(t *testing.T) {
	for _, tt := range headerTests {
		r, _ := NewRequest("GET", "http://example.com", Headers(map[string]string{tt.k: tt.v}))
		v := r.Header.Get(tt.k)
		assert.Equal(t, v, tt.v)
	}
}

func TestAll(t *testing.T) {
	params := map[string]string{
		"b": "b",
		"a": "c",
	}
	req, _ := NewRequest("get", "http://simple.org/path/?b=a",
		Params(params),
		Auth(authTests[0].username, authTests[0].password))
	assert.Equal(t, req.URL.String(), "http://simple.org/path/?a=c&b=b")
	username, password, ok := req.BasicAuth()
	assert.Equal(t, ok, authTests[0].ok)
	assert.Equal(t, username, authTests[0].username)
	assert.Equal(t, password, authTests[0].password)
}

var jsonTests = struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}{
	"name",
	21,
}

func TestJSON(t *testing.T) {
	req, _ := NewRequest("POST", "http://httpbin.org/post", JSON(jsonTests))
	b, _ := json.Marshal(jsonTests)
	reqE, _ := http.NewRequest("POST", "http://httpbin.org/post", bytes.NewReader(b))
	assert.Equal(t, req.Body, reqE.Body)
	assert.Equal(t, req.ContentLength, reqE.ContentLength)
}

func TestFile(t *testing.T) {
	filename := "testdata/file4upload"
	req, err := NewRequest("POST", "http://example.com", FileContent(filename))
	assert.Equal(t, err, nil)

	f, err := os.Open(filename)
	assert.Equal(t, err, nil)

	fb, err := ioutil.ReadAll(f)
	assert.Equal(t, err, nil)

	assert.Equal(t, req.ContentLength, int64(len(fb)))
	b, err := ioutil.ReadAll(req.Body)
	assert.Equal(t, err, nil)
	assert.Equal(t, b, fb)
}

func TestCookies(t *testing.T) {
	c := map[string]string{
		"Key1": "Value1",
		"Key2": "Value2",
	}
	req, err := NewRequest("POST", "http://example.com", Cookies(c))
	assert.Equal(t, err, nil)

	cookies := req.Cookies()
	assert.Equal(t, len(c), len(cookies))
	for _, cookie := range cookies {
		value, ok := c[cookie.Name]
		assert.Equal(t, ok, true)
		assert.Equal(t, value, cookie.Value)
	}
}

func TestData(t *testing.T) {
	f1 := map[string]string{
		"a": "1",
	}

	f2 := map[string]string{
		"a": "2",
		"b": "2",
	}

	req, err := NewRequest("POST", "http://example.com", Data(f1), Data(f2))
	assert.Equal(t, err, nil)

	for k, v := range f2 {
		assert.Equal(t, req.PostForm.Get(k), v)
	}

}

func TestBody(t *testing.T) {
	body := strings.NewReader("test string body")
	req, err := NewRequest("GET", "http://example.com", Body(body))
	assert.Nil(t, err)

	req2, err := http.NewRequest("GET", "http://example.com", body)
	assert.Nil(t, err)

	assert.Equal(t, req2.Body, req.Body)
	b, err := req.GetBody()
	assert.Nil(t, err)
	b2, err := req2.GetBody()
	assert.Nil(t, err)

	assert.Equal(t, b2, b)
}

func TestMultipartForm(t *testing.T) {
	f, err := os.Open("testdata/file4upload")
	assert.Nil(t, err)
	form := MultipartForm(map[string]io.Reader{
		"field1": strings.NewReader("value1"),
		"field2": strings.NewReader("value2"),
		"file1":  f,
	})

	req, err := NewRequest("POST", "http://example.com", form)
	assert.Nil(t, err)

	body, err := req.GetBody()
	assert.Nil(t, err)
	n, err := io.ReadAll(body)
	assert.Nil(t, err)
	assert.Equal(t, int64(len(n)), req.ContentLength)
	_ = body.Close()
}
