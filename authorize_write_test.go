/*
 * Copyright © 2015-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * @author		Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @copyright 	2015-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @license 	Apache-2.0
 *
 */

package fosite_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	. "github.com/ory/fosite"
	. "github.com/ory/fosite/internal"
)

func TestWriteAuthorizeResponse(t *testing.T) {
	oauth2 := &Fosite{}
	header := http.Header{}
	ctrl := gomock.NewController(t)
	rw := NewMockResponseWriter(ctrl)
	ar := NewMockAuthorizeRequester(ctrl)
	resp := NewMockAuthorizeResponder(ctrl)
	defer ctrl.Finish()

	for k, c := range []struct {
		setup  func()
		expect func()
	}{
		{
			setup: func() {
				redir, _ := url.Parse("https://foobar.com/?foo=bar")
				ar.EXPECT().GetRedirectURI().Return(redir)
				ar.EXPECT().GetResponseMode().Return(ResponseModeDefault)
				resp.EXPECT().GetParameters().Return(url.Values{})
				resp.EXPECT().GetHeader().Return(http.Header{})

				rw.EXPECT().Header().Return(header).Times(2)
				rw.EXPECT().WriteHeader(http.StatusSeeOther)
			},
			expect: func() {
				assert.Equal(t, http.Header{
					"Location":      []string{"https://foobar.com/?foo=bar"},
					"Cache-Control": []string{"no-store"},
					"Pragma":        []string{"no-cache"},
				}, header)
			},
		},
		{
			setup: func() {
				redir, _ := url.Parse("https://foobar.com/?foo=bar")
				ar.EXPECT().GetRedirectURI().Return(redir)
				ar.EXPECT().GetResponseMode().Return(ResponseModeFragment)
				resp.EXPECT().GetParameters().Return(url.Values{"bar": {"baz"}})
				resp.EXPECT().GetHeader().Return(http.Header{})

				rw.EXPECT().Header().Return(header).Times(2)
				rw.EXPECT().WriteHeader(http.StatusSeeOther)
			},
			expect: func() {
				assert.Equal(t, http.Header{
					"Location":      []string{"https://foobar.com/?foo=bar#bar=baz"},
					"Cache-Control": []string{"no-store"},
					"Pragma":        []string{"no-cache"},
				}, header)
			},
		},
		{
			setup: func() {
				redir, _ := url.Parse("https://foobar.com/?foo=bar")
				ar.EXPECT().GetRedirectURI().Return(redir)
				ar.EXPECT().GetResponseMode().Return(ResponseModeQuery)
				resp.EXPECT().GetParameters().Return(url.Values{"bar": {"baz"}})
				resp.EXPECT().GetHeader().Return(http.Header{})

				rw.EXPECT().Header().Return(header).Times(2)
				rw.EXPECT().WriteHeader(http.StatusSeeOther)
			},
			expect: func() {
				expectedUrl, _ := url.Parse("https://foobar.com/?foo=bar&bar=baz")
				actualUrl, err := url.Parse(header.Get("Location"))
				assert.Nil(t, err)
				assert.Equal(t, expectedUrl.Query(), actualUrl.Query())
				assert.Equal(t, "no-cache", header.Get("Pragma"))
				assert.Equal(t, "no-store", header.Get("Cache-Control"))
			},
		},
		{
			setup: func() {
				redir, _ := url.Parse("https://foobar.com/?foo=bar")
				ar.EXPECT().GetRedirectURI().Return(redir)
				ar.EXPECT().GetResponseMode().Return(ResponseModeFragment)
				resp.EXPECT().GetParameters().Return(url.Values{"bar": {"b+az ab"}})
				resp.EXPECT().GetHeader().Return(http.Header{"X-Bar": {"baz"}})

				rw.EXPECT().Header().Return(header).Times(2)
				rw.EXPECT().WriteHeader(http.StatusSeeOther)
			},
			expect: func() {
				assert.Equal(t, http.Header{
					"X-Bar":         {"baz"},
					"Location":      {"https://foobar.com/?foo=bar#bar=b%2Baz+ab"},
					"Cache-Control": []string{"no-store"},
					"Pragma":        []string{"no-cache"},
				}, header)
			},
		},
		{
			setup: func() {
				redir, _ := url.Parse("https://foobar.com/?foo=bar")
				ar.EXPECT().GetRedirectURI().Return(redir)
				ar.EXPECT().GetResponseMode().Return(ResponseModeQuery)
				resp.EXPECT().GetParameters().Return(url.Values{"bar": {"b+az"}, "scope": {"a b"}})
				resp.EXPECT().GetHeader().Return(http.Header{"X-Bar": {"baz"}})

				rw.EXPECT().Header().Return(header).Times(2)
				rw.EXPECT().WriteHeader(http.StatusSeeOther)
			},
			expect: func() {
				expectedUrl, err := url.Parse("https://foobar.com/?foo=bar&bar=b%2Baz&scope=a+b")
				assert.Nil(t, err)
				actualUrl, err := url.Parse(header.Get("Location"))
				assert.Nil(t, err)
				assert.Equal(t, expectedUrl.Query(), actualUrl.Query())
				assert.Equal(t, "no-cache", header.Get("Pragma"))
				assert.Equal(t, "no-store", header.Get("Cache-Control"))
				assert.Equal(t, "baz", header.Get("X-Bar"))
			},
		},
		{
			setup: func() {
				redir, _ := url.Parse("https://foobar.com/?foo=bar")
				ar.EXPECT().GetRedirectURI().Return(redir)
				ar.EXPECT().GetResponseMode().Return(ResponseModeFragment)
				resp.EXPECT().GetParameters().Return(url.Values{"scope": {"api:*"}})
				resp.EXPECT().GetHeader().Return(http.Header{"X-Bar": {"baz"}})

				rw.EXPECT().Header().Return(header).Times(2)
				rw.EXPECT().WriteHeader(http.StatusSeeOther)
			},
			expect: func() {
				assert.Equal(t, http.Header{
					"X-Bar":         {"baz"},
					"Location":      {"https://foobar.com/?foo=bar#scope=api%3A%2A"},
					"Cache-Control": []string{"no-store"},
					"Pragma":        []string{"no-cache"},
				}, header)
			},
		},
		{
			setup: func() {
				redir, _ := url.Parse("https://foobar.com/?foo=bar#bar=baz")
				ar.EXPECT().GetRedirectURI().Return(redir)
				ar.EXPECT().GetResponseMode().Return(ResponseModeFragment)
				resp.EXPECT().GetParameters().Return(url.Values{"qux": {"quux"}})
				resp.EXPECT().GetHeader().Return(http.Header{})

				rw.EXPECT().Header().Return(header).Times(2)
				rw.EXPECT().WriteHeader(http.StatusSeeOther)
			},
			expect: func() {
				assert.Equal(t, http.Header{
					"Location":      {"https://foobar.com/?foo=bar#qux=quux"},
					"Cache-Control": []string{"no-store"},
					"Pragma":        []string{"no-cache"},
				}, header)
			},
		},
		{
			setup: func() {
				redir, _ := url.Parse("https://foobar.com/?foo=bar")
				ar.EXPECT().GetRedirectURI().Return(redir)
				ar.EXPECT().GetResponseMode().Return(ResponseModeFragment)
				resp.EXPECT().GetParameters().Return(url.Values{"state": {"{\"a\":\"b=c&d=e\"}"}})
				resp.EXPECT().GetHeader().Return(http.Header{})

				rw.EXPECT().Header().Return(header).Times(2)
				rw.EXPECT().WriteHeader(http.StatusSeeOther)
			},
			expect: func() {
				assert.Equal(t, http.Header{
					"Location":      {"https://foobar.com/?foo=bar#state=%7B%22a%22%3A%22b%3Dc%26d%3De%22%7D"},
					"Cache-Control": []string{"no-store"},
					"Pragma":        []string{"no-cache"},
				}, header)
			},
		},
		{
			setup: func() {
				redir, _ := url.Parse("https://foobar.com/?foo=bar")
				ar.EXPECT().GetRedirectURI().Return(redir)
				ar.EXPECT().GetResponseMode().Return(ResponseModeFormPost)
				resp.EXPECT().GetHeader().Return(http.Header{"X-Bar": {"baz"}})
				resp.EXPECT().GetParameters().Return(url.Values{"code": {"poz65kqoneu"}, "state": {"qm6dnsrn"}})

				rw.EXPECT().Header().Return(header).AnyTimes()
				rw.EXPECT().Write(gomock.Any()).AnyTimes()
			},
			expect: func() {
				assert.Equal(t, "text/html;charset=UTF-8", header.Get("Content-Type"))
			},
		},
	} {
		t.Logf("Starting test case %d", k)
		c.setup()
		oauth2.WriteAuthorizeResponse(rw, ar, resp)
		c.expect()
		header = http.Header{}
		t.Logf("Passed test case %d", k)
	}
}
