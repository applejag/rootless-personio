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

package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var rawFlags = struct {
	method   string
	data     string
	json     string
	headers  []string
	formData []string
}{}

var rawCmd = &cobra.Command{
	Use:   "raw </url/path?query=param>",
	Short: "Send a raw HTTP request to the API",
	Long: `Send a raw HTTP request to the API
as a logged in user, and print the resulting JSON data.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		urlArg := args[0]
		baseURL, err := getBaseURL(urlArg)
		if err != nil {
			return err
		}
		if baseURL != "" {
			cfg.BaseURL = baseURL
		}

		client, err := newLoggedInClient()
		if err != nil {
			return err
		}

		method := http.MethodGet

		body, err := getDataFromRawFlags()
		if err != nil {
			return err
		}
		if body != nil {
			method = http.MethodPost
			defer body.Close()
		}

		if rawFlags.method != "" {
			method = rawFlags.method
		}

		req, err := http.NewRequest(method, urlArg, body)
		if err != nil {
			return err
		}
		for _, header := range rawFlags.headers {
			key, value, ok := strings.Cut(header, ":")
			if !ok {
				return fmt.Errorf(`invalid header, expected "Key: value", got %q`, header)
			}
			// Adds header while maintaining the capitalization
			req.Header[key] = append(req.Header[key], strings.TrimPrefix(value, " "))
		}

		var resp *http.Response
		var respErr error
		switch {
		case rawFlags.json != "":
			resp, respErr = client.RawJSON(req)
		case len(rawFlags.formData) > 0:
			resp, respErr = client.RawForm(req)
		default:
			resp, respErr = client.Raw(req)
		}
		if resp != nil {
			defer resp.Body.Close()
			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("read response body: %w", err)
			}

			if responseIsJSON(resp) {
				var model any
				if err := json.Unmarshal(respBody, &model); err != nil {
					return err
				}
				if err := printOutputJSONOrYAML(model); err != nil {
					return err
				}
			} else {
				fmt.Println(string(respBody))
			}
		}
		return respErr
	},
}

func responseIsJSON(resp *http.Response) bool {
	contentType := resp.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	switch {
	case err != nil:
		return false
	case mediaType == "application/json":
		return true
	case strings.HasPrefix(mediaType, "application/") &&
		strings.HasSuffix(mediaType, "+json"):
		return true
	default:
		return false
	}
}

func init() {
	rootCmd.AddCommand(rawCmd)

	rawCmd.Flags().StringVarP(&rawFlags.method, "method", "X", rawFlags.method, `Request method to use (default "POST" if with --data flag, otherwise "GET")`)
	rawCmd.Flags().StringVarP(&rawFlags.data, "data", "d", rawFlags.data, `Request body ("@filename" for reading from file, or "@-" from STDIN)`)
	rawCmd.Flags().StringVar(&rawFlags.json, "json", rawFlags.json, `Request JSON body, same as --data, but sends and expects "Content-Type: application/json"`)
	rawCmd.Flags().StringArrayVarP(&rawFlags.headers, "header", "H", nil, `Add custom headers to request (format "Key: value")`)
	rawCmd.Flags().StringArrayVarP(&rawFlags.formData, "form", "F", nil, `Add multipart MIME data, and send "Content-Type: application/x-www-form-urlencoded" (format "key=value")`)
}

func getBaseURL(urlArg string) (string, error) {
	u, err := url.Parse(urlArg)
	if err != nil {
		return "", fmt.Errorf("parse positional arg as URL: %w", err)
	}
	if u.Host == "" {
		return "", nil
	}
	u.Path = ""
	return u.String(), nil
}

func getDataFromRawFlags() (io.ReadCloser, error) {
	// read from --json
	jsonData, err := getDataFlagReader(rawFlags.json)
	if err != nil || jsonData != nil {
		return jsonData, err
	}
	// read from --data
	binaryData, err := getDataFlagReader(rawFlags.data)
	if err != nil || binaryData != nil {
		return binaryData, err
	}
	// read from --form
	if len(rawFlags.formData) > 0 {
		var values url.Values
		for _, pair := range rawFlags.formData {
			key, value, ok := strings.Cut(pair, "=")
			if !ok {
				return nil, fmt.Errorf(`invalid form data, expected "key=value", got %q`, pair)
			}
			values.Add(key, value)
		}
		return io.NopCloser(strings.NewReader(values.Encode())), nil
	}
	return nil, nil
}

func getDataFlagReader(dataFlag string) (io.ReadCloser, error) {
	if dataFlag == "" {
		return nil, nil
	}
	if dataFlag[0] == '@' {
		path := dataFlag[1:]

		if path == "-" {
			return os.Stdin, nil
		}

		file, err := os.Open(path)
		return file, err
	}
	strReader := strings.NewReader(dataFlag)
	return io.NopCloser(strReader), nil
}
