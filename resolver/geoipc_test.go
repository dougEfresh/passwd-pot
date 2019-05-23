// Copyright © 2019.  Douglas Chimento <dchimento@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resolver

import "testing"

func TestLookupGeo(t *testing.T) {
	g := GeoClient{
		"http://localhost:8080",
	}
	i, err := g.GetLocationForAddr("4.2.2.2")
	if err != nil {
		t.Fatalf("Could not lookup ip %s", err)
	}
	if i.IP == "" {
		t.Fatalf("Bad response %s", i)
	}
}
