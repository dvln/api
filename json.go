// Copyright © 2015 Erik Brady <brady@dvln.org>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// The dvln/api/json.go module is for routines that might be useful
// for manipulating json beyond (or wrapping) the Go standard lib,
// also useful for "combining" JSON fields/data/errors from various
// packages that can then be dumped as a whole from anywhere.

package api

// for imports the goal is to use very little outside the std lib,
// note that str and cast have no dependencies outside the std lib
// (exception: cast testing file which uses 'testify')
import (
	"bytes"
	"dvln/lib/out"
	"encoding/json"
	"os"

	"github.com/mgutz/str"
	"github.com/spf13/cast"
)

var jsonIndentLevel = 2
var jsonPrefix = ""
var jsonRaw = false

// JSONIndentLevel can be used to get the current indentation level for each
// "step" in PrettyJSON() output (defaults to 2 currently)
func JSONIndentLevel() int {
	return jsonIndentLevel
}

// SetJSONIndentLevel can be used to change the indentation level for each
// "step" in pretty JSOn output being formatted via PrettyJSON()
func SetJSONIndentLevel(level int) {
	jsonIndentLevel = level
}

// JSONPrefix can be used to get the current prefix used for any JSON string
// being formatted via the PrettyJSON() routine
func JSONPrefix() string {
	return jsonPrefix
}

// SetJSONPrefix can be used to change the string prefix for any JSON string
// being formatted via the PrettyJSON() routine.
func SetJSONPrefix(pfx string) {
	jsonPrefix = pfx
}

// JSONRaw can be used to determine if we're in raw JSON output mode (true)
// or not, true means the PrettyJSON() routine will do nothing
func JSONRaw() bool {
	return jsonRaw
}

// SetJSONRaw can be used to change the indentation level for each
// "step" in pretty JSOn output being formatted via PrettyJSON()
// being formatted via the PrettyJSON() routine.
func SetJSONRaw(b bool) {
	jsonRaw = b
}

// PrettyJSON pretty prints JSON data, provide the data and that can be followed
// by two optional arguments, a prefix string and an indent level (both of which
// are strings).  If neither is provided then no prefix used and indent of two
// spaces is the default (see cfgfile:jsonprefix, cfgfile:jsonindent and the
// related DVLN_JSONPREFIX, DVLN_JSONINDENT to adjust indentation and prefix
// as well as cfgfile:jsonraw and DVLN_JSONRAW for skipping pretty printing)
func PrettyJSON(b []byte, fmt ...string) (string, error) {
	if jsonRaw {
		// if there's an override to say pretty JSON is not desired, honor it,
		// Feature: this could be changed to specifically remove carriage
		//          returns and shorten output around {} and :'s and such (?)
		return cast.ToString(b), nil
	}
	prefix := jsonPrefix
	indent := str.Pad("", " ", jsonIndentLevel)
	if len(fmt) == 1 {
		prefix = fmt[0]
	} else if len(fmt) == 2 {
		prefix = fmt[0]
		indent = fmt[1]
	}
	var out bytes.Buffer
	err := json.Indent(&out, b, prefix, indent)
	return cast.ToString(out.Bytes()) + "\n", err
}

// GetJSONString takes the various things needed from a DVLN api call and
// combines everything into a passable JSON string (pretty or not depending
// upon settings) and returns that representation to the caller.
func GetJSONString(apiVer string, context string, kind string, verbosity string, fields []string, items []interface{}) string {
	// FIXME: eriknow: ideally errors should result in a JSON "string" that is a
	//        normal JSON error being returned... see ~/dvln.txt for example
	if apiVer == "" {
		// In case the API version couldn't be passed, last ditch try
		apiVer = os.Getenv("PKG_API_APIVER")
		if apiVer == "" {
			out.Fatalln("No API version was available to GetJSONString(), failing")
		}
	}
	apiRoot := NewRoot(apiVer, context).SetAPIItems(kind, verbosity, fields, items)
	j, err := json.Marshal(apiRoot)
	if err != nil {
		out.Issuef("Failed to convert %s (%s) items to JSON: %s\n", context, kind, err)
		return ""
	}
	var str string
	// put in indentation and formatting, can turn that off as well
	// if desired via the "jsonraw" glob (viper) setting
	str, err = PrettyJSON(j)
	if err != nil {
		out.Issueln("Unable to beautify JSON output:", err)
		return ""
	}
	return str
}
