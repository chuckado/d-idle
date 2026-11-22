package main

import (
	"regexp"
	"strconv"
)

var paceVirtual = regexp.MustCompile(`"msg":"pace".*?"virtual":(\d+)`)

// lastVirtualMin is the virtual minutes the bot's last `pace` log line reports (0 when there is none).
func lastVirtualMin(log []byte) float64 {
	all := paceVirtual.FindAllSubmatch(log, -1)
	if len(all) == 0 {
		return 0
	}
	ns, _ := strconv.ParseFloat(string(all[len(all)-1][1]), 64)
	return ns / 6e10
}

// restHours is the budget left for the second leg of a restart run, in hours (never below one virtual minute).
func restHours(totalMin, doneMin float64) string {
	left := totalMin - doneMin
	if left < 1 {
		left = 1
	}
	return strconv.FormatFloat(left/60, 'f', -1, 64)
}
