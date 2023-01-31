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

package personio_test

import (
	"log"
	"os"

	"github.com/jilleJr/rootless-personio/pkg/personio"
)

func Example() {
	client, err := personio.New("https://example.personio.de")
	if err != nil {
		log.Fatalln("Error creating client:", err)
	}

	email := os.Getenv("PERSONIO_EMAIL")
	password := os.Getenv("PERSONIO_PASS")
	if email == "" || password == "" {
		log.Fatalln("Must set env var PERSONIO_EMAIL and PERSONIO_PASS")
	}

	if client.Login(email, password); err != nil {
		log.Fatalln("Error logging in:", err)
	}

	log.Println("Logged in as employee ID:", client.EmployeeID)

	currentEmployee, err := client.GetMyEmployeeData()
	if err != nil {
		log.Fatalln("Error fetching employee:", err)
	}

	log.Printf("Welcome, %s %s, %s of the %s team!",
		currentEmployee.FirstName,
		currentEmployee.LastName,
		currentEmployee.Position,
		currentEmployee.Department,
	)
}
