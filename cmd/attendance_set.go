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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/jilleJr/rootless-personio/pkg/personio"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/typ.v4/slices"
)

var attendanceSetFlags = struct {
	file string
}{}

var attendanceSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Sets attendance periods",
	Long: `Sets (updates and replaces) attendance periods on multiple days.

You provide a list of periods, and all days mentioned in that list will be
updated by this command.

The input is provided by JSON objects in a file as specified with the --file flag,
(or when set to "--file -", piping in JSON through STDIN).
The input should be a stream of Personio attendance periods. Example:

    {
      "start": "2023-01-18T08:00:00Z",
      "end": "2023-01-18T12:00:00Z",
      "comment": "Work before lunch",
      "period_type": "work"
    }
    {
      "start": "2023-01-18T12:00:00Z",
      "end": "2023-01-18T13:00:00Z",
      "comment": "Lunch break",
      "period_type": "break"
    }
    {
      "start": "2023-01-18T13:00:00Z",
      "end": "2023-01-18T17:00:00Z",
      "comment": "Work after lunch",
      "period_type": "work"
    }

It is incorrect to provide a JSON array with the elements.
If you have a JSON array, you can convert it to a stream via jq like so:

    jq '.[]' my-file.json
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var file io.ReadCloser = os.Stdin
		if attendanceSetFlags.file != "-" {
			var err error
			file, err = os.Open(attendanceSetFlags.file)
			if err != nil {
				return err
			}
		}
		defer file.Close()

		var periods []personio.Period
		dec := json.NewDecoder(file)
		for {
			var p personio.Period
			err := dec.Decode(&p)
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				return fmt.Errorf("read periods: %w", err)
			}
			dur := p.End.Sub(p.Start)

			log.Debug().
				Str("type", string(p.PeriodType)).
				Time("start", p.Start).
				Time("end", p.End).
				Str("dur", dur.Truncate(time.Second).String()).
				Str("comment", p.GetComment()).
				Msg("Read attendance period.")

			if dur < cfg.MinimumPeriodDuration {
				log.Warn().
					Str("type", string(p.PeriodType)).
					Time("start", p.Start).
					Time("end", p.End).
					Str("dur", dur.Truncate(time.Second).String()).
					Str("comment", p.GetComment()).
					Str("minimumDuration", cfg.MinimumPeriodDuration.String()).
					Msg("Skipping period because it has a too short duration.")
				continue
			}

			periods = append(periods, p)
		}

		if len(periods) == 0 {
			return errors.New("missing periods, please provide JSON objects via STDIN or --file")
		}

		periodsPerDay := slices.GroupBy(periods, func(p personio.Period) string {
			return p.Start.Format("2006-01-02")
		})
		slices.SortFunc(periodsPerDay, func(a, b slices.Grouping[string, personio.Period]) bool {
			return a.Key < b.Key
		})

		client, err := newLoggedInClient()
		if err != nil {
			return err
		}

		type PerDay struct {
			Day     string            `json:"day"`
			Periods []personio.Period `json:"periods"`
		}
		var printableGroups []PerDay

		for _, group := range periodsPerDay {
			err = client.SetAttendance(group.Values[0].Start, group.Values)
			if err != nil {
				return err
			}
			log.Info().
				Str("day", group.Key).
				Int("periods", len(group.Values)).
				Msg("Successfully updated attendance for day.")
			printableGroups = append(printableGroups, PerDay{
				Day:     group.Key,
				Periods: group.Values,
			})
		}

		return printOutputJSONOrYAML(map[string]any{
			"groups": printableGroups,
		})
	},
}

func init() {
	attendanceCmd.AddCommand(attendanceSetCmd)

	attendanceSetCmd.Flags().StringVarP(&attendanceSetFlags.file, "file", "f", "", `Attendance periods JSON file, "-" means STDIN`)
	attendanceSetCmd.MarkFlagFilename("file", "json")
	attendanceSetCmd.MarkFlagRequired("file")
}
