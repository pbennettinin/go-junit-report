package main

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"
)

type Result int

const (
	PASS Result = iota
	FAIL
)

type Report struct {
	Packages []Package
}

type Package struct {
	Name  string
	Time  int
	Tests []Test
}

type Test struct {
	Name   string
	Time   int
	Result Result
	Output []string
}

var (
	regexStart  = regexp.MustCompile(`^=== RUN:? (.+)$`)
	regexStatus = regexp.MustCompile(`^--- (PASS|FAIL|SKIP): (.+) \(([0-9.]+) ?(?:s|seconds|ms|us)?\)$`)
	regexResult = regexp.MustCompile(`^(ok|FAIL)\s+(.+)\s([0-9.]+)s$`)
)

func Parse(r io.Reader) (*Report, error) {
	reader := bufio.NewReader(r)

	report := &Report{make([]Package, 0)}

	// keep track of tests we find
	tests := make([]Test, 0)

	// current test
	var test *Test

	// parse lines
	for {
		l, _, err := reader.ReadLine()
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		line := string(l)

		if matches := regexStart.FindStringSubmatch(line); len(matches) == 2 {
			// start of a new test
			if test != nil {
				tests = append(tests, *test)
			}

			test = &Test{
				Name:   matches[1],
				Result: FAIL,
				Output: make([]string, 0),
			}
		} else if matches := regexResult.FindStringSubmatch(line); len(matches) == 4 {
			// all tests in this package are finished
			if test != nil {
				tests = append(tests, *test)
				test = nil
			}

			report.Packages = append(report.Packages, Package{
				Name:  matches[2],
				Time:  parseTime(matches[3]),
				Tests: tests,
			})

			tests = make([]Test, 0)
		} else if test != nil {
			if matches := regexStatus.FindStringSubmatch(line); len(matches) == 4 {
				// test status
				if matches[1] == "PASS" || matches[1] == "SKIP" {
					test.Result = PASS
				} else {
					test.Result = FAIL
				}

				test.Name = matches[2]
				test.Time = parseTime(matches[3]) * 10
			} else {
				if line == "PASS" || line == "FAIL" || strings.HasPrefix(line, "exit status") {
					// status for end of tests - don't add as output into prior test
				} else {
					// log ALL test output
					test.Output = append(test.Output, line)
				}
			}
		}
	}

	return report, nil
}

func parseTime(time string) int {
	t, err := strconv.Atoi(strings.Replace(time, ".", "", -1))
	if err != nil {
		return 0
	}
	return t
}
