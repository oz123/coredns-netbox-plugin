// Copyright 2020 Oz Tiram <oz.tiram@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package netbox

import (
	//"fmt"
	"testing"
	"time"

	"github.com/coredns/caddy"
	"github.com/stretchr/testify/assert"
)

// TestSetup tests the various things that should be parsed by setup.
func TestSetup(t *testing.T) {
	// tests to run
	tests := []struct {
		msg     string
		input   string
		wantErr bool
	}{
		{
			"minimal valid config",
			"netbox {\nurl example.org\ntoken foobar\nlocalCacheDuration 10s\n}\n",
			false,
		},
		{
			"empty config",
			"netbox {}\n",
			true,
		},
		{
			"invalid localCacheDuration",
			"netbox {\nurl example.org\ntoken foobar\nlocalCacheDuration Wrong\n}\n",
			true,
		},
		{
			"config with ttl",
			"netbox {\nurl example.org\ntoken foobar\nlocalCacheDuration 10s\nttl 3600s\n}\n",
			false,
		},
		{
			"config with invalid ttl",
			"netbox {\nurl example.org\ntoken foobar\nlocalCacheDuration 10s\nttl INVALID\n}\n",
			true,
		},
		{
			"config with fallthrough (all)",
			"netbox {\nurl example.org\ntoken foobar\nlocalCacheDuration 10s\nfallthrough\n}\n",
			false,
		},
		{
			"config with fallthrough (one domain)",
			"netbox {\nurl example.org\ntoken foobar\nlocalCacheDuration 10s\nfallthrough example.org\n}\n",
			false,
		},
		{
			"config with fallthrough (multiple domains)",
			"netbox {\nurl example.org\ntoken foobar\nlocalCacheDuration 10s\nfallthrough example.org example.net\n}\n",
			false,
		},
	}

	// run tests
	for _, tt := range tests {
		c := caddy.NewTestController("dns", tt.input)
		err := setup(c)
		if tt.wantErr {
			assert.Error(t, err, tt.msg)
		} else {
			assert.Nil(t, err, tt.msg)
		}
	}
}

func TestSetupWithDuration(t *testing.T) {
	c := caddy.NewTestController("dns", `netbox { url example.org\n token foobar\n localCacheDuration 10s }`)
	nb, _ := newNetBox(c)
	assert.Equal(t, nb.CacheDuration, time.Second*10, "Duration not set properly")
}
