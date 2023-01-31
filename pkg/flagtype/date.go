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

package flagtype

import (
	"time"

	"github.com/spf13/pflag"
)

type Date time.Time

// ensure it implements the interface
var _ pflag.Value = &Date{}

// Time is a helper function to return the [time.Time] representation.
func (d Date) Time() time.Time {
	return time.Time(d)
}

// IsZero returns true when this date is set to it's zero value: 0001-01-01
func (d Date) IsZero() bool {
	return time.Time(d) == time.Time{}
}

// String implements [fmt.Stringer] and [pflag.Value].
//
// Used by cobra when showing the default value of a flag.
func (d Date) String() string {
	if d.IsZero() {
		return ""
	}
	return d.Time().Format("2006-01-02")
}

// Set implements [pflag.Value].
//
// Used by cobra when setting the new value for a flag.
func (d *Date) Set(value string) error {
	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		return err
	}
	*d = Date(t.UTC())
	return nil
}

// Type implements [pflag.Value].
//
// Used by cobra when rendering the list of flags and their types.
func (d Date) Type() string {
	return "date"
}
