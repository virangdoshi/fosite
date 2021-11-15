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

package oauth2

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/fosite"
	"github.com/ory/fosite/internal"
	"github.com/ory/fosite/token/jwt"
)

func TestIntrospectJWT(t *testing.T) {
	strat := &DefaultJWTStrategy{
		JWTStrategy: &jwt.RS256JWTStrategy{
			PrivateKey: internal.MustRSAKey(),
		},
	}

	v := &StatelessJWTValidator{
		JWTStrategy:   strat,
		ScopeStrategy: fosite.HierarchicScopeStrategy,
	}

	for k, c := range []struct {
		description string
		token       func() string
		expectErr   error
		scopes      []string
	}{
		{
			description: "should fail because jwt is expired",
			token: func() string {
				jwt := jwtExpiredCase(fosite.AccessToken)
				token, _, err := strat.GenerateAccessToken(nil, jwt)
				assert.NoError(t, err)
				return token
			},
			expectErr: fosite.ErrTokenExpired,
		},
		{
			description: "should pass because scope was granted",
			token: func() string {
				jwt := jwtValidCase(fosite.AccessToken)
				jwt.GrantedScope = []string{"foo", "bar"}
				token, _, err := strat.GenerateAccessToken(nil, jwt)
				assert.NoError(t, err)
				return token
			},
			scopes: []string{"foo"},
		},
		{
			description: "should fail because scope was not granted",
			token: func() string {
				jwt := jwtValidCase(fosite.AccessToken)
				token, _, err := strat.GenerateAccessToken(nil, jwt)
				assert.NoError(t, err)
				return token
			},
			scopes:    []string{"foo"},
			expectErr: fosite.ErrInvalidScope,
		},
		{
			description: "should fail because signature is invalid",
			token: func() string {
				jwt := jwtValidCase(fosite.AccessToken)
				token, _, err := strat.GenerateAccessToken(nil, jwt)
				assert.NoError(t, err)
				parts := strings.Split(token, ".")
				require.Len(t, parts, 3, "%s - %v", token, parts)
				dec, err := base64.RawURLEncoding.DecodeString(parts[1])
				assert.NoError(t, err)
				s := strings.Replace(string(dec), "peter", "piper", -1)
				parts[1] = base64.RawURLEncoding.EncodeToString([]byte(s))
				return strings.Join(parts, ".")
			},
			expectErr: fosite.ErrTokenSignatureMismatch,
		},
		{
			description: "should pass",
			token: func() string {
				jwt := jwtValidCase(fosite.AccessToken)
				token, _, err := strat.GenerateAccessToken(nil, jwt)
				assert.NoError(t, err)
				return token
			},
		},
	} {
		t.Run(fmt.Sprintf("case=%d:%v", k, c.description), func(t *testing.T) {
			if c.scopes == nil {
				c.scopes = []string{}
			}

			areq := fosite.NewAccessRequest(nil)
			_, err := v.IntrospectToken(nil, c.token(), fosite.AccessToken, areq, c.scopes)

			if c.expectErr != nil {
				require.EqualError(t, err, c.expectErr.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, "peter", areq.Session.GetSubject())
			}
		})
	}
}

func BenchmarkIntrospectJWT(b *testing.B) {
	strat := &DefaultJWTStrategy{
		JWTStrategy: &jwt.RS256JWTStrategy{
			PrivateKey: internal.MustRSAKey(),
		},
	}

	v := &StatelessJWTValidator{
		JWTStrategy: strat,
	}

	jwt := jwtValidCase(fosite.AccessToken)
	token, _, err := strat.GenerateAccessToken(nil, jwt)
	assert.NoError(b, err)
	areq := fosite.NewAccessRequest(nil)

	for n := 0; n < b.N; n++ {
		_, err = v.IntrospectToken(nil, token, fosite.AccessToken, areq, []string{})
	}

	assert.NoError(b, err)
}
