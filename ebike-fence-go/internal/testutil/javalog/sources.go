package javalog

import (
	"fmt"
	"os"
	"path/filepath"
)

// DefaultSourceLogs are Java ebike-fence log files used for fixture extraction.
var DefaultSourceLogs = []string{
	"ebike-fence-1.log",
	"ebike-fence-2.log",
	"ebike-fence-3.log",
}

// LoadDefaultSourceEntries parses all default log files under root.
func LoadDefaultSourceEntries(root string) ([]Entry, error) {
	var all []Entry
	for _, name := range DefaultSourceLogs {
		p := filepath.Join(root, name)
		entries, err := ParseFile(p)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "skip %s: not found\n", name)
				continue
			}
			return nil, fmt.Errorf("parse %s: %w", name, err)
		}
		fmt.Fprintf(os.Stderr, "parsed %s: %d entries\n", name, len(entries))
		all = append(all, entries...)
	}
	if len(all) == 0 {
		return nil, fmt.Errorf("no log entries under %s", root)
	}
	return all, nil
}
