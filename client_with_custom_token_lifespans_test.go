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

package fosite

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultClientWithCustomTokenLifespans(t *testing.T) {
	clc := &DefaultClientWithCustomTokenLifespans{
		DefaultClient: &DefaultClient{},
	}

	assert.Equal(t, clc.GetTokenLifespans(), (*ClientLifespanConfig)(nil))

	require.Equal(t, time.Minute*42, GetEffectiveLifespan(clc, GrantTypeImplicit, IDToken, time.Minute*42))

	customLifespan := 36 * time.Hour
	clc.SetTokenLifespans(&ClientLifespanConfig{ImplicitGrantIDTokenLifespan: &customLifespan})
	assert.NotEqual(t, clc.GetTokenLifespans(), nil)

	require.Equal(t, customLifespan, GetEffectiveLifespan(clc, GrantTypeImplicit, IDToken, time.Minute*42))
	var _ ClientWithCustomTokenLifespans = clc
}
