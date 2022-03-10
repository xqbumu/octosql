/*
Copyright 2021 The Vitess Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v2

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConvertMySQLVersion(t *testing.T) {
	testcases := []struct {
		version        string
		commentVersion string
		error          string
	}{{
		version:        "5.7.9",
		commentVersion: "50709",
	}, {
		version:        "0008.08.9",
		commentVersion: "80809",
	}, {
		version:        "5.7.9, Vitess - 10.0.1",
		commentVersion: "50709",
	}, {
		version:        "8.1 Vitess - 10.0.1",
		commentVersion: "80100",
	}, {
		version: "Vitess - 10.0.1",
		error:   "MySQL version not correctly setup - Vitess - 10.0.1.",
	}, {
		version:        "5.7.9.22",
		commentVersion: "50709",
	}}

	for _, tcase := range testcases {
		t.Run(tcase.version, func(t *testing.T) {
			output, err := convertMySQLVersionToCommentVersion(tcase.version)
			if tcase.error != "" {
				require.EqualError(t, err, tcase.error)
			} else {
				require.NoError(t, err)
				require.Equal(t, tcase.commentVersion, output)
			}
		})
	}
}
