// Copyright © 2017 Douglas Chimento <dchimento@gmail.com>
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

package potdb

import (
	"os"
	"strings"
	"testing"

	"github.com/dougEfresh/passwd-pot/log"
)

const testdsn string = "postgres://postgres:@127.0.0.1/?sslmode=disable"

var db DB
var tester *potDB

var logger = log.DefaultLogger(os.Stdout)

func init() {
	var err error
	db, err = Open(testdsn)

	if err != nil {
		logger.Errorf("DB error %s ", err)
	}
	tester = &potDB{
		db:    db.Get(),
		mysql: !strings.Contains(testdsn, "postgres"),
	}
}
func TestParams(t *testing.T) {
	if db == nil {
		t.Fatal("No db")
	}

	params := tester.replaceParams("? ? ? ? ?")
	if params != "$1 $2 $3 $4 $5" {
		t.Fatalf("Params are wrong %s", params)
	}
	tester.mysql = true
	params = tester.replaceParams("? ? ? ? ?")
	if params != "? ? ? ? ?" {
		t.Fatalf("Params are wrong %s", params)
	}
}
