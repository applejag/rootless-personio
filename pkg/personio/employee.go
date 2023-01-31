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
	"errors"
	"fmt"
	"net/http"
)

type Employee struct {
	ID           int             `json:"id"`
	FirstName    string          `json:"first_name"`
	LastName     string          `json:"last_name"`
	Position     string          `json:"position"`
	Department   string          `json:"department"`
	Office       string          `json:"office"`
	Team         string          `json:"team"`
	AccessRights map[string]bool `json:"access_rights"`
}

type EmployeeProfileImage struct {
	Small    string `json:"small"`
	Medium   string `json:"medium"`
	Large    string `json:"large"`
	Original string `json:"original"`
}

type EmployeeTab struct {
	Name     string `json:"name"`
	Route    string `json:"route"`
	Label    string `json:"label"`
	IsActive bool   `json:"isActive"`
}

func (c *Client) GetMyEmployeeData() (*Employee, error) {
	if c.EmployeeID == 0 {
		return nil, errors.New("no employee ID stored in client")
	}
	return c.GetEmployeeData(c.EmployeeID)
}

func (c *Client) GetEmployeeData(id int) (*Employee, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/employee-header-bff/%d", id), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	resp, err := c.RawJSON(req)
	if err != nil {
		return nil, err
	}
	return ParseResponseJSON[*Employee](resp)
}
