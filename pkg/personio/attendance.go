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
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/jilleJr/rootless-personio/pkg/util"
	"github.com/rs/zerolog/log"
)

const timeDateOnlyLayout = "2006-01-02"

type AttendanceCalendar struct {
	AttendanceRights         map[string]bool                  `json:"attendance_rights"`
	EmployeeWorkingSchedules struct{}                         `json:"employee_working_schedules"`
	AttendanceDays           Data[[]CalendarDay]              `json:"attendance_days"`
	AttendancePeriods        Data[[]CalendarAttendancePeriod] `json:"attendance_periods"`
	OvertimeItems            struct{}                         `json:"overtime_items"`
	AttendanceAlerts         struct{}                         `json:"attendance_alerts"`
	AbsencePeriods           Data[[]CalendarAbsencePeriod]    `json:"absence_periods"`
	Holidays                 Data[[]CalendarHoliday]          `json:"holidays"`
}

type CalendarDay struct {
	ID         uuid.UUID             `json:"id"` // ex: "d5bb4b32-c499-4f79-a534-93481505bd60"
	Attributes CalendarDayAttributes `json:"attributes"`
}

type CalendarDayAttributes struct {
	BreakMin    int    `json:"break_min"`    // Duration of breaks in minutes
	DurationMin int    `json:"duration_min"` // Duration of attendance in minutes
	Status      string `json:"status"`       // ex: "empty"
	Day         string `json:"day"`          // ex: "2023-01-20"
}

type CalendarAttendancePeriod struct {
	ID         uuid.UUID                          `json:"id"` // ex: "bc1edc0c-44ef-467f-89a0-10d0733efec5"
	Attributes CalendarAttendancePeriodAttributes `json:"attributes"`
}

type CalendarAttendancePeriodAttributes struct {
	AttendanceDayID uuid.UUID `json:"attendance_day_id"` // ex: "81954d73-0b0d-4053-a5dc-937bdd62f9f7"
	Comment         *string   `json:"comment"`           // ex: ""
	End             string    `json:"end"`               // ex: "2023-01-18T17:00:00Z"
	LegacyBreakMin  int       `json:"legacy_break_min"`  // ex: 0
	PeriodType      string    `json:"period_type"`       // ex: "work"
	ProjectID       *int      `json:"project_id"`
	Start           string    `json:"start"` // ex: "2023-01-18T13:00:00Z"
}

type CalendarAbsencePeriod struct {
	ID                         string `json:"id"`   // ex: "123456789"
	Name                       string `json:"name"` // ex: "Paid vacation"
	TracksOvertime             bool   `json:"tracks_overtime"`
	MeasurementUnit            string `json:"measurement_unit"` // ex: "day"
	StartDate                  string `json:"start_date"`       // ex: "2022-12-22"
	StartTime                  string `json:"start_time"`       // ex: "2022-12-22 00:00:00"
	EndDate                    string `json:"end_date"`         // ex: "2022-12-28"
	EndTime                    string `json:"end_time"`         // ex: "2022-12-29 00:00:00"
	EffectiveDurationInMinutes *int   `json:"effective_duration_in_minutes"`
	HalfDayStart               bool   `json:"half_day_start"`
	HalfDayEnd                 bool   `json:"half_day_end"`
}

type CalendarHoliday struct {
	HalfDay             bool   `json:"half_day"`
	HolidayCalendarName string `json:"holiday_calendar_name"` // ex: "DE (Hamburg) Feiertage CompanyName"
	ID                  int    `json:"id"`                    // ex: 123456
	Name                string `json:"name"`                  // ex: "2. Weihnachtstag"
	Date                string `json:"date"`                  // ex: "2022-12-26"
}

type Data[M any] struct {
	Data M `json:"data"`
}

func (c *Client) GetMyAttendanceCalendar(startDate, endDate time.Time) (*AttendanceCalendar, error) {
	return c.GetAttendanceCalendar(c.EmployeeID, startDate, endDate)
}

func (c *Client) GetAttendanceCalendar(employeeID int, startDate, endDate time.Time) (*AttendanceCalendar, error) {
	if err := c.assertLoggedIn(); err != nil {
		return nil, err
	}

	queryParams := url.Values{}
	queryParams.Set("start_date", startDate.Format(timeDateOnlyLayout))
	queryParams.Set("end_date", endDate.Format(timeDateOnlyLayout))

	req, err := http.NewRequest("GET", fmt.Sprintf(
		"/svc/attendance-bff/attendance-calendar/%d?%s",
		employeeID, queryParams.Encode()), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.RawJSON(req)
	if err != nil {
		return nil, err
	}

	return ParseResponseJSON[*AttendanceCalendar](resp)
}

type Period struct {
	ID         uuid.UUID  `json:"id"`          // ex: "46365bc8-482a-41b2-8d36-68491140edd9"
	PeriodType PeriodType `json:"period_type"` // ex: "work"
	Comment    *string    `json:"comment"`     // ex: ""
	ProjectID  *int       `json:"project_id"`  // ex: null
	Start      time.Time  `json:"start"`       // ex: "2023-01-18T08:00:00Z"
	End        time.Time  `json:"end"`         // ex: "2023-01-18T12:00:00Z"

	// Required by the HTTP API, but seemingly unused
	LegacyBreakMin int `json:"legacy_break_min"` // ex: 0
}

func (p Period) GetComment() string {
	if p.Comment == nil {
		return ""
	}
	return *p.Comment
}

func (p Period) GetProjectID() int {
	if p.ProjectID == nil {
		return 0
	}
	return *p.ProjectID
}

type PeriodType string

const (
	PeriodTypeWork  PeriodType = "work"
	PeriodTypeBreak PeriodType = "break"
)

func (c *Client) SetAttendance(date time.Time, periods []Period) error {
	if err := c.assertLoggedIn(); err != nil {
		return err
	}

	for i := range periods {
		if periods[i].ID == uuid.Nil {
			periods[i].ID = uuid.New()
		}
		periods[i].Start = periods[i].Start.Truncate(time.Second).UTC()
		periods[i].End = periods[i].End.Truncate(time.Second).UTC()
		if periods[i].PeriodType == "" {
			periods[i].PeriodType = PeriodTypeWork
		}
	}

	body, err := json.Marshal(map[string]any{
		"employee_id": c.EmployeeID,
		"periods":     periods,
	})
	if err != nil {
		return err
	}
	bodyReader := bytes.NewReader(body)

	dayID, err := c.GetOrNewDayUUID(date)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, "/api/v1/attendances/days/"+dayID.String(), bodyReader)
	if err != nil {
		return err
	}

	resp, err := c.RawJSON(req)
	if err != nil {
		return err
	}

	// Currently don't care about the response
	_, err = ParseResponseJSON[any](resp)
	return err
}

// GetOrNewDayUUID will either lookup a day's ID (from cache or by querying
// the API), or generate a new ID and store this new ID in cache.
//
// After the remote lookup to the API, the client caches which days in the same
// month that has undefined IDs.
func (c *Client) GetOrNewDayUUID(date time.Time) (uuid.UUID, error) {
	id, err := c.GetDayUUID(date)
	if err != nil {
		return uuid.Nil, fmt.Errorf("get day UUID: %w", err)
	}
	if id != nil {
		return *id, nil
	}
	newID := uuid.New()
	dateString := date.Format(timeDateOnlyLayout)
	c.dayIDCache[dateString] = &newID
	log.Debug().Str("day", dateString).Stringer("uuid", newID).
		Msg("Randomized new UUID for day.")
	return newID, nil
}

// GetDayUUID will lookup a day's ID (from cache or by querying the API),
// or nil if it is undefined.
//
// The Personio API want the client to generate the IDs, so an undefined day ID
// means you are free to generate your own ID.
//
// After the remote lookup to the API, the client caches which days in the same
// month that has undefined IDs.
func (c *Client) GetDayUUID(date time.Time) (*uuid.UUID, error) {
	dateString := date.Format(timeDateOnlyLayout)
	// Cache contains nil values on "known to be undefined day IDs"
	if id, ok := c.dayIDCache[dateString]; ok {
		return id, nil
	}
	startDate, endDate := util.TimeFullMonth(date)
	cal, err := c.GetMyAttendanceCalendar(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("get days for range %s-%s: %w",
			startDate.Format(timeDateOnlyLayout),
			endDate.Format(timeDateOnlyLayout),
			err)
	}

	c.cacheDayIDs(cal.AttendanceDays.Data, startDate, endDate)
	return c.dayIDCache[dateString], nil
}

func (c *Client) cacheDayIDs(days []CalendarDay, startDate, endDate time.Time) {
	// Cache known days
	for _, day := range days {
		// must clone the var so we don't take ref of the for loop var
		id := day.ID
		c.dayIDCache[day.Attributes.Day] = &id
		log.Debug().Str("day", day.Attributes.Day).Stringer("uuid", id).
			Msg("Cached existing UUID for day.")
	}

	// Set unknown days
	loopEnd := endDate.Add(24 * time.Hour)
	for date := startDate; date.Before(loopEnd); date = date.Add(24 * time.Hour) {
		dateString := date.Format(timeDateOnlyLayout)
		if _, ok := c.dayIDCache[dateString]; !ok {
			c.dayIDCache[dateString] = nil
		}
	}
}
