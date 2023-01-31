// SPDX-FileCopyrightText: 2023 Kalle Fagerberg
// SPDX-FileCopyrightText: 2022 Risk.Ident GmbH <contact@riskident.com>
//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify it
// under the terms of the GNU General Public License as published by the
// Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for
// more details.
//
// You should have received a copy of the GNU General Public License along
// with this program.  If not, see <http://www.gnu.org/licenses/>.

package util

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/mattn/go-isatty"
)

var camelCaseReplacer = strings.NewReplacer(
	"ID", "Id",
	"URL", "Url",
	"HTTP", "Http",
	"JSON", "Json",
	"JQ", "Jq",
	"YAML", "Yaml",
	"YQ", "Yq",
	"GitHub", "Github",
	"PR", "Pr",
	"API", "Api",
	"PEM", "Pem",
	"DER", "Pem",
	"RSA", "Rsa",
)

// ToCamelCase is a very stupid implementation for converting
// PascalCase to camelCase.
//
// NOTE: If we need to convert user-provided strings to camelCase,
// then we should replace this with the community strcase package.
func ToCamelCase(s string) string {
	if len(s) == 0 {
		return s
	}
	s = camelCaseReplacer.Replace(s)
	b := []byte(s)
	b[0] = byte(unicode.ToLower(rune(b[0])))
	return string(b)
}

func PrettyPath(s string) string {
	path := filepath.Clean(s)
	if workingDir, err := os.Getwd(); err == nil {
		if rel, err := filepath.Rel(workingDir, path); err == nil {
			if !strings.HasPrefix(rel, fmt.Sprintf("..%c", filepath.Separator)) {
				return rel
			}
		}
	}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		trimmed := strings.TrimPrefix(path, home)
		if len(trimmed) != len(path) {
			return filepath.Join("~", trimmed)
		}
	}
	return path
}

func ColorizeJSON(data []byte) ([]byte, error) {
	args := []string{"."}
	if isatty.IsTerminal(os.Stdout.Fd()) {
		args = append(args, "--color-output")
	}

	jq := exec.Command("jq", args...)
	jq.Stdin = bytes.NewReader(data)
	colorized, err := jq.Output()
	if err != nil {
		return nil, err
	}
	return colorized, nil
}

func ColorizeYAML(data []byte) ([]byte, error) {
	args := []string{".", "-"}
	if isatty.IsTerminal(os.Stdout.Fd()) {
		args = append(args, "--colors")
	}

	yq := exec.Command("yq", args...)
	yq.Stdin = bytes.NewReader(data)
	colorized, err := yq.Output()
	if err != nil {
		return nil, err
	}
	return colorized, nil
}
