// SPDX-FileCopyrightText: 2022 Jonas Riedel
// SPDX-FileCopyrightText: 2023 Kalle Fagerberg
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

package personio

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"gopkg.in/typ.v4"
)

func TestCacheDayIDs(t *testing.T) {
	var knownDays = []CalendarDay{
		{
			ID:         uuid.MustParse("00000000-0000-0000-0000-000000000001"),
			Attributes: CalendarDayAttributes{Day: "2023-01-01"},
		},
		{
			ID:         uuid.MustParse("00000000-0000-0000-0000-000000000005"),
			Attributes: CalendarDayAttributes{Day: "2023-01-05"},
		},
		{
			ID:         uuid.MustParse("00000000-0000-0000-0000-000000000010"),
			Attributes: CalendarDayAttributes{Day: "2023-01-10"},
		},
		{
			ID:         uuid.MustParse("00000000-0000-0000-0000-000000000011"),
			Attributes: CalendarDayAttributes{Day: "2023-01-11"},
		},
	}
	startDate := mustParseTime(t, "2006-01-02", "2023-01-01")
	endDate := mustParseTime(t, "2006-01-02", "2023-01-31")

	client := &Client{dayIDCache: make(map[string]*uuid.UUID)}
	client.cacheDayIDs(knownDays, startDate, endDate)

	cache := client.dayIDCache

	wantLen := 31
	if len(cache) != wantLen {
		t.Errorf("want len %d, got %d", wantLen, len(cache))
	}

	wantMap := map[string]*uuid.UUID{
		"2023-01-01": typ.Ref(uuid.MustParse("00000000-0000-0000-0000-000000000001")),
		"2023-01-02": nil,
		"2023-01-03": nil,
		"2023-01-04": nil,
		"2023-01-05": typ.Ref(uuid.MustParse("00000000-0000-0000-0000-000000000005")),
		"2023-01-06": nil,
		"2023-01-07": nil,
		"2023-01-08": nil,
		"2023-01-09": nil,
		"2023-01-10": typ.Ref(uuid.MustParse("00000000-0000-0000-0000-000000000010")),
		"2023-01-11": typ.Ref(uuid.MustParse("00000000-0000-0000-0000-000000000011")),
		"2023-01-12": nil,
		"2023-01-13": nil,
		"2023-01-14": nil,
		"2023-01-15": nil,
		"2023-01-16": nil,
		"2023-01-17": nil,
		"2023-01-18": nil,
		"2023-01-19": nil,
		"2023-01-20": nil,
		"2023-01-21": nil,
		"2023-01-22": nil,
		"2023-01-23": nil,
		"2023-01-24": nil,
		"2023-01-25": nil,
		"2023-01-26": nil,
		"2023-01-27": nil,
		"2023-01-28": nil,
		"2023-01-29": nil,
		"2023-01-30": nil,
		"2023-01-31": nil,
	}

	if len(cache) != len(wantMap) {
		t.Errorf("want len %d, got %d", len(wantMap), len(cache))
	} else {
		t.Logf("ok len of %d", len(wantMap))
	}

	for key, want := range wantMap {
		got, ok := cache[key]
		switch {
		case !ok:
			t.Errorf("want key %q, but key was missing", key)
			continue
		case want == nil && got != nil:
			t.Errorf("want key %q to be nil, but was %s", key, got)
		case want != nil && got == nil:
			t.Errorf("want key %q to be %s, but was nil", key, want)
		case want != nil && got != nil && *want != *got:
			t.Errorf("want key %q to be %s, but was %s", key, want, got)
		}
	}

	for key, got := range cache {
		_, ok := wantMap[key]
		if !ok {
			t.Errorf("do not want key %q, but was %s", key, got)
		}
	}
}

func mustParseTime(t *testing.T, layout, value string) time.Time {
	t.Helper()
	tim, err := time.Parse(layout, value)
	if err != nil {
		t.Fatalf("mustParseTime: time.Parse(%q, %q): %s", layout, value, err)
	}
	return tim
}
