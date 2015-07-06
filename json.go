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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/dvln/cast"
	"github.com/dvln/str"
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

// EscapeCtrl escapes control chars in a string so JSON likes em
func EscapeCtrl(ctrl []byte) (esc []byte) {
	u := []byte(`\u0000`)
	for i, ch := range ctrl {
		if ch <= 31 {
			if esc == nil {
				esc = append(make([]byte, 0, len(ctrl)+len(u)), ctrl[:i]...)
			}
			esc = append(esc, u...)
			hex.Encode(esc[len(esc)-2:], ctrl[i:i+1])
			continue
		}
		if esc != nil {
			esc = append(esc, ch)
		}
	}
	if esc == nil {
		return ctrl
	}
	return esc
}

// FatalJSONMsg is for cases where Marshal is failing so we need
// some JSON we can dump on the output... if we get to this level then
// what we're generating is a valid JSON error basically (shouldn't happen)
func FatalJSONMsg(apiVer string, msg Msg) string {
	cleanMsg := EscapeCtrl([]byte(msg.Message))
	rawJSON := fmt.Sprintf("{ \"apiVersion\":\"%s\", \"id\": -1, \"error\": { \"message\": \"%s\", \"code\": %d, \"level\": \"%s\" } }", apiVer, cleanMsg, msg.Code, msg.Level)
	output, err := PrettyJSON([]byte(rawJSON))
	if err != nil {
		output = rawJSON
	}
	return output
}

// GetJSONOutput takes the various things needed from a DVLN api call and
// combines everything into a passable JSON string (pretty or not depending
// upon settings) and returns that representation to the caller.
func GetJSONOutput(apiVer string, context string, kind string, verbosity string, fields []string, items []interface{}) (string, bool) {
	var j []byte
	var err error
	var output, rawJSON string
	var errMsg, warnMsg Msg
	fatalErr := false

	if apiVer == "" {
		// In case the API version couldn't be passed, last ditch try
		apiVer = os.Getenv("PKG_API_APIVER")
		if apiVer == "" {
			apiVer = "?.?"
			errMsg.Message = "No valid JSON API version is available"
			errMsg.Code = 1001
			errMsg.Level = "FATAL"
			fatalErr = true
		}
	}
	apiRoot := newAPIData(apiVer, context)
	if errMsg.Message == "" && storedFatal.Message != "" {
		errMsg = storedFatal
		fatalErr = true
	} else if errMsg.Message == "" && storedWarning.Message != "" {
		warnMsg = storedWarning
	}
	if errMsg.Message == "" {
		// if no errors so far then add in our items and 'data' details
		apiRoot.SetAPIItems(kind, verbosity, fields, items)
		if warnMsg.Message != "" {
			apiRoot.Warning = warnMsg
		}
	} else {
		// otherwise indicate issue and encode that into JSON
		apiRoot.ID = -1
		apiRoot.Error = errMsg
	}
	j, err = json.Marshal(apiRoot)
	if err != nil {
		if errMsg.Message == "" {
			errMsg.Message = "Unable to marshal basic JSON API string"
			errMsg.Code = 1002
			errMsg.Level = "FATAL"
			fatalErr = true
		}
		// hack: hard code some JSON and return an error... shouldn't happen
		rawJSON = FatalJSONMsg(apiVer, errMsg)
		return rawJSON, fatalErr
	}
	// put in indentation and formatting, can turn that off as well
	// if desired via the "jsonraw" globs (viper) setting
	output, err = PrettyJSON(j)
	if err != nil {
		warnMsg.Message = fmt.Sprintf("Unable to beautify JSON output: %s", err)
		warnMsg.Code = 1003
		warnMsg.Level = "ISSUE"
		apiRoot.Warning = warnMsg
		j, err = json.Marshal(apiRoot)
		// if 1st marshal ok but pretty failed, add warning to JSON and if basic
		// re-Marshal fails for any reason "bump" to a FATAL error, unlikely:
		if err != nil {
			// not a warning any more, scale it up to fatal error
			warnMsg.Level = "FATAL"
			fatalErr = true
			rawJSON = FatalJSONMsg(apiVer, warnMsg)
			return rawJSON, fatalErr
		}
		// retry pretty probably won't work again, if not just use raw json
		output, err = PrettyJSON(j)
		if err != nil {
			output = cast.ToString(j)
		}
	}
	// Return the output (typically), fatalErr is false if we get to here
	return output, fatalErr
}
