/* simple config file reader
   the format is toml like
     # comment
     [scope]
     name = val
     ..
*/

package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"ednode/pkg/slog"
)

// config var name-val pair type
type nvPair map[string]string

// parsed and stored config data itself
var nvData map[string]*nvPair

// parsed line
type parsedLine struct {
	scope string
	name  string
	value string
}

// read and parse config file
func Read(path string) error {

	// init config data storage
	nvData = make(map[string]*nvPair)

	// open config file
	fp, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fp.Close()

	scanner := bufio.NewScanner(fp)
	scanner.Split(bufio.ScanLines)

	// read config file
	for lineNo := 1; scanner.Scan(); lineNo++ {

		line := prepareLine(scanner.Text())
		if line == "" {
			continue
		}

		pl, err := parseLine(line)
		if err != nil {
			return fmt.Errorf("line %d: %s (near '%.16s ...')", lineNo, err, line)
		}

		// a line is a scope line, nothing yet to save
		if pl == nil {
			continue
		}

		slog.Debug(9, "configured: [%s] '%s' => '%v'", pl.scope, pl.name, pl.value)

		nvSet(pl.scope, pl.name, pl.value)
	}

	return nil
}

// prepare config file line for parsing
func prepareLine(line string) string {
	// skip empty lines and comments
	l := strings.TrimSpace(line)
	if len(l) < 1 || l[0] == '#' {
		return ""
	}
	return l
}

// parse prepared config line
var currScope string

func parseLine(line string) (*parsedLine, error) {
	lineLen := len(line)

	// no config line can be shorter than 3 chars (i.e. A=B or [G])
	if lineLen < 3 {
		return nil, fmt.Errorf("too short line")
	}

	// got a scope defining line: remember scope name and return
	if line[0] == '[' && line[lineLen-1] == ']' {
		currScope = strings.TrimSpace(line[1 : lineLen-1])
		if currScope == "" {
			return nil, fmt.Errorf("invalid scope name")
		}
		return nil, nil
	}

	// no scope defined yet
	if currScope == "" {
		return nil, fmt.Errorf("expression without scope")
	}

	// try to get name=value pair here
	tokens := strings.SplitN(line, "=", 2)
	if len(tokens) != 2 {
		return nil, fmt.Errorf("can not parse")
	}

	name := strings.TrimSpace(tokens[0])
	if len(name) < 1 {
		return nil, fmt.Errorf("param name missed")
	}

	value := strings.TrimSpace(tokens[1])
	if len(value) < 1 {
		return nil, fmt.Errorf("param value missed")
	}

	return &parsedLine{
		scope: currScope,
		name:  name,
		value: value,
	}, nil
}

// set (overriding) name-value pair in aprropriate scope
func nvSet(scope string, name string, value string) {
	nv := make(nvPair)
	nv[name] = value
	nvData[scope] = &nv
}

// get named string value from specified scope
func ValStr(scope string, name string) (string, error) {
	if nvp, ok := nvData[scope]; ok {
		if val, ok := nvPair(*nvp)[name]; ok {
			return val, nil
		}
	}
	return "", fmt.Errorf("'%s' is not found in '%s'", name, scope)
}

// get names int32 value from specified scope
func ValInt32(scope string, name string) (int32, error) {
	if val, err := ValStr(scope, name); err != nil {
		return 0, err
	} else if i, err := strconv.ParseInt(val, 10, 0); err != nil {
		return 0, err
	} else {
		return int32(i), nil
	}
}

// get names float value from specified scope
func ValFloat32(scope string, name string) (float32, error) {
	if val, err := ValStr(scope, name); err != nil {
		return 0, err
	} else if i, err := strconv.ParseFloat(val, 32); err != nil {
		return 0, err
	} else {
		return float32(i), nil
	}
}
