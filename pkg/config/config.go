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

// Package config contains the configuration data structs and flag types
// used in the command line tool.
package config

import (
	"reflect"
	"time"

	"github.com/invopop/jsonschema"
	"github.com/jilleJr/rootless-personio/pkg/util"
)

// Config is the full configuration file.
type Config struct {
	// BaseURL is the URL to your Personio instance.
	// This can be with or without the trailing slash.
	//
	// The program with later append paths like /login/index
	// and /api/v1/attendances/periods when invoking its HTTP
	// requests.
	//
	// Any query parameters and fragments will get removed.
	BaseURL string `yaml:"baseUrl" jsonschema:"oneof_type=string;null" jsonschema_extras:"format=uri"`
	Auth    Auth

	// MinimumPeriodDuration is the duration for which attendance periods that
	// are shorter than will get skipped when creating or updating attendance.
	//
	// The value is a Go duration, which allows values like:
	// - 30s
	// - 12m30s
	// - 2h12m30s
	MinimumPeriodDuration time.Duration `yaml:"minimumPeriodDuration" jsonschema:"type=string"`

	// Output is the format of the command line results.
	// This controls the format of the single command line
	// result output written to STDOUT.
	Output OutFormat
	Log    Log
}

// Auth contains configs for how the program should authenticate
// with Personio.
type Auth struct {
	// Email is your account's login email address.
	Email string `jsonschema:"oneof_type=string;null" jsonschema_extras:"format=email"`
	// Password is your account's login password.
	Password string `jsonschema:"oneof_type=string;null"`

	// CSRFToken is provided by this program when it fails to
	// log in due to them detecting login via new device. You then need to
	// run the program again but with the CSRF (Cross-Site-Request-Forgery)
	// token and email token.
	CSRFToken string `yaml:"csrfToken,omitempty" jsonschema:"oneof_type=string;null"`
	// EmailToken is sent by Personio to your email when it fails to
	// log in due to them detecting login via new device. You then need to
	// run the program again but with the CSRF (Cross-Site-Request-Forgery)
	// token and email token.
	EmailToken string `yaml:"emailToken,omitempty" jsonschema:"oneof_type=string;null"`
}

// Log contains configs for the command line logging, which compared
// to the command line output, loggin is written to STDERR and contains
// small status reports, and is mostly used for debugging.
type Log struct {
	// Format is the way the program formats its logging line. The
	// "pretty" option is meant for humans and is colored, while the
	// "json" option is meant for easier parsing in logging management
	// systems like for example Kibana or Splunk.
	Format LogFormat
	// Level is the severity level to filter logs on, where "trace"
	// is the lowest logging/severity level, and "panic" is the
	// highest. The program will only log messages that are equal
	// severity or higher than this value. You can also set this
	// to "disabled" to turn of logging.
	Level LogLevel
}

type jsonSchemaInterface interface {
	JSONSchema() *jsonschema.Schema
}

// Schema returns the JSON schema for the [Config] struct.
//
// Supports optionally supplying the source directory path,
// which can point to the Go module directory, and will then
// use the comments from the source code as descriptions in
// the resulting schema.
func Schema(sourceDir string) *jsonschema.Schema {
	r := new(jsonschema.Reflector)
	r.KeyNamer = util.ToCamelCase
	r.Namer = func(t reflect.Type) string {
		return util.ToCamelCase(t.Name())
	}
	r.RequiredFromJSONSchemaTags = true
	if sourceDir != "" {
		r.AddGoComments("github.com/jilleJr/rootless-personio", sourceDir)
	}
	s := r.Reflect(&Config{})
	s.ID = "https://github.com/jilleJr/rootless-personio/raw/main/personio.schema.json"
	return s
}
