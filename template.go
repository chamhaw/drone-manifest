package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/aymerick/raymond"
)

func init() {
	raymond.RegisterHelpers(funcs)
}

// Render parses and executes a template, returning the results in string format.
func Render(template string, payload interface{}) (s string, err error) {
	u, err := url.Parse(template)
	if err == nil {
		switch u.Scheme {
		case "http", "https":
			res, err := http.Get(template)
			if err != nil {
				return s, err
			}
			defer res.Body.Close()
			out, err := ioutil.ReadAll(res.Body)
			if err != nil {
				return s, err
			}
			template = string(out)

		default:
			out, err := ioutil.ReadFile(u.Path)
			if err != nil {
				return s, err
			}
			template = string(out)
		}
	}

	return raymond.Render(template, payload)
}

// RenderTrim parses and executes a template, returning the results in string
// format. The result is trimmed to remove left and right padding and newlines
// that may be added unintentially in the template markup.
func RenderTrim(template string, playload interface{}) (string, error) {
	out, err := Render(template, playload)
	return strings.Trim(out, " \n"), err
}

var funcs = map[string]interface{}{
	"uppercase":      strings.ToUpper,
	"lowercase":      strings.ToLower,
	"trimPrefix":     strings.TrimPrefix,
	"trimSuffix":     strings.TrimSuffix,
	"quote":          strconv.Quote,
	"join":           strings.Join,
	"uppercasefirst": uppercaseFirst,
	"duration":       toDuration,
	"datetime":       toDatetime,
	"success":        isSuccess,
	"failure":        isFailure,
	"truncate":       truncate,
	"urlencode":      urlencode,
	"since":          since,
}

func uppercaseFirst(s string) string {
	a := []rune(s)
	a[0] = unicode.ToUpper(a[0])
	s = string(a)
	return s
}

func toDuration(started, finished float64) string {
	return fmt.Sprintln(time.Duration(finished-started) * time.Second)
}

func toDatetime(timestamp float64, layout, zone string) string {
	if len(zone) == 0 {
		return time.Unix(int64(timestamp), 0).Format(layout)
	}
	loc, err := time.LoadLocation(zone)
	if err != nil {
		return time.Unix(int64(timestamp), 0).Local().Format(layout)
	}
	return time.Unix(int64(timestamp), 0).In(loc).Format(layout)
}

func isSuccess(conditional bool, options *raymond.Options) string {
	if !conditional {
		return options.Inverse()
	}

	switch options.ParamStr(0) {
	case "success":
		return options.Fn()
	default:
		return options.Inverse()
	}
}

func isFailure(conditional bool, options *raymond.Options) string {
	if !conditional {
		return options.Inverse()
	}

	switch options.ParamStr(0) {
	case "failure", "error", "killed":
		return options.Fn()
	default:
		return options.Inverse()
	}
}

func truncate(s string, len int) string {
	if utf8.RuneCountInString(s) <= len {
		return s
	}
	runes := []rune(s)
	return string(runes[:len])
}

func urlencode(options *raymond.Options) string {
	return url.QueryEscape(options.Fn())
}

func since(start int64) string {
	// NOTE: not using `time.Since()` because the fractional second component
	// will give us something like "40m12.917523438s" vs "40m12s". We lose
	// some precision, but the format is much more readable.
	now := time.Unix(time.Now().Unix(), 0)
	return fmt.Sprintln(now.Sub(time.Unix(start, 0)))
}
