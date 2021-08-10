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
// Make sure you also test for parse errors.
func TestSetup(t *testing.T) {
	c := caddy.NewTestController("dns", `netbox { url example.org\n token foobar\n localCacheDuration 10s }`)
	if err := setup(c); err != nil {
		t.Fatalf("Expected no errors, but got: %v", err)
	}

	c = caddy.NewTestController("dns", `netbox {}`)
	if err := setup(c); err == nil {
		t.Fatalf("Expected errors, but got: %v", err)
	}
}

func TestSetupWithWrongDuration(t *testing.T) {
	c := caddy.NewTestController("dns", `netbox { url example.org\n token foobar\n localCacheDuration Wrong }`)
	_, err := newNetBox(c)
	assert.Error(t, err, "Expected error")
}

func TestSetupWithDuration(t *testing.T) {
	c := caddy.NewTestController("dns", `netbox { url example.org\n token foobar\n localCacheDuration 10s }`)
	nb, _ := newNetBox(c)
	assert.Equal(t, nb.CacheDuration, time.Second*10, "Duration not set properly")
}
