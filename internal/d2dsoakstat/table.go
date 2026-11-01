package d2dsoakstat

import (
	"bufio"
	"os"
	"strings"
)

// LogTable is the leveler skill table the bot reported on its minute lines (last `table=` value, text or JSON log),
// or "" for a log from before change fixstatic or an unreadable file.
func LogTable(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	var last string
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		line := sc.Text()
		if !strings.Contains(line, "minute") {
			continue
		}
		var kv map[string]string
		switch {
		case strings.HasPrefix(line, "{"):
			kv = jsonKV(line)
		case strings.Contains(line, "msg=minute "):
			kv = textKV(line)
		}
		if kv != nil && kv["msg"] == "minute" && kv["table"] != "" {
			last = kv["table"]
		}
	}
	return last
}
