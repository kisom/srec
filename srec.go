// Package srec returns Motorola S-Record dumps. Currently, it only
// supports S19 and S37 records, but not S28 records.
package srec

import (
	"fmt"
	"strings"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func encodeRecord(stype uint8, raw []byte) string {
	out := fmt.Sprintf("S%d%x\n", stype, raw)
	return strings.ToUpper(out)
}
