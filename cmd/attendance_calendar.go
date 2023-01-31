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

package cmd

import (
	"time"

	"github.com/jilleJr/rootless-personio/pkg/flagtype"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var attendanceCalendarFlags = struct {
	startDate flagtype.Date
	endDate   flagtype.Date
}{}

var attendanceCalendarCmd = &cobra.Command{
	Use:   "calendar",
	Short: "Show the calendar of your attendance",
	RunE: func(cmd *cobra.Command, args []string) error {
		startDate := attendanceCalendarFlags.startDate.Time()
		endDate := attendanceCalendarFlags.endDate.Time()
		if !cmd.Flag("start").Changed {
			year, month, _ := time.Now().Date()
			startDate = time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
		}
		if !cmd.Flag("end").Changed {
			year, month, _ := time.Now().Date()
			endDate = time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC)
		}

		log.Debug().
			Time("start", startDate).
			Time("end", endDate).
			Msg("Date range.")
		client, err := newLoggedInClient()
		if err != nil {
			return err
		}
		cal, err := client.GetMyAttendanceCalendar(startDate, endDate)
		if err != nil {
			return err
		}

		return printOutputJSONOrYAML(cal)
	},
}

func init() {
	attendanceCmd.AddCommand(attendanceCalendarCmd)

	attendanceCalendarCmd.Flags().VarP(&attendanceCalendarFlags.startDate, "start", "s", "Start date to show (default first day this month)")
	attendanceCalendarCmd.Flags().VarP(&attendanceCalendarFlags.endDate, "end", "e", "End date to show (default first day this month)")
}
