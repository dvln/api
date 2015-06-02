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

// The dvln/api/api.go module is for routines to "build up" an API structure
// that can then be mapped to the desired format (eg: JSON) and dumped for
// the user.  This is a hack at this point.

package api

// for imports the goal is to use very little outside the std lib,
// note that str and cast have no dependencies outside the std lib
// (exception: cast testing file which uses 'testify')

/* FIXME: stuff to log in a usage log... but not always returned
          for basic JSON queries at this point
type JSONLog struct {
	startTime
	endTime
}
*/

// Root is a structure mapping to the "root" API settings (currently the
// API is dumped in JSON format).  If field aren't provided then they will
// not be shown but one must have APIVersion defined (and ID will come back
// as 0 if not later set meaning "success" as it maps to the exit value of
// the tool essentially)
type Root struct {
	APIVersion string      `json:"apiVersion"`
	Context    string      `json:"context,omitempty"`
	ID         int         `json:"id"`
	Data       interface{} `json:"data,omitempty"`
	Error      interface{} `json:"error,omitempty"`
}

// NewRoot basically sets up a new API "root" structure which contains the
// API version, a given context (eg: "dvlnGlobs", "dvlnGet") and a deafult
// ID of 0... along with empty pointers to Data and Error to be fleshed out
// by the caller (data: items: [..] or error: {errdata})
func NewRoot(apiVersion string, context string) *Root {
	var root Root
	root.APIVersion = apiVersion
	root.Context = context
	return &root
}

// SetAPIItems will take a more detailed "kind" of items (eg: 'env' or 'cfg'
// for Globs data), an optional verbosity (use "" to skip), the fields maps to
// the fields available within each item included and the items themselves
// which must be an array of interface{} for this to fly.
func (r *Root) SetAPIItems(kind string, verbosity string, fields []string, items []interface{}) *Root {
	type jsonData struct {
		Kind             string        `json:"kind,omitempty"`
		Verbosity        string        `json:"verbosity,omitempty"`
		Fields           []string      `json:"fields,omitempty"`
		TotalItems       int           `json:"totalItems,omitempty"`
		StartIndex       int           `json:"startIndex,omitempty"`
		CurrentItemCount int           `json:"currentItemCount,omitempty"`
		Items            []interface{} `json:"items,omitempty"`
	}
	var data jsonData
	data.Kind = kind
	data.Verbosity = verbosity
	data.Fields = fields
	length := len(items)
	data.TotalItems = length
	data.StartIndex = 1
	data.CurrentItemCount = length
	data.Items = items
	r.Data = &data
	return r
}