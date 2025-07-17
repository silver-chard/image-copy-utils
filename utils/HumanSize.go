package utils

import "fmt"

var unitArray = []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}

func HumanSize(size uint64) string {
	rem, i := uint64(0), 0
	for size >= 1000 && i < len(unitArray)-1 {
		rem = size % 1000
		size /= 1000
		i++
	}
	return fmt.Sprintf("%d.%02d %s", size, (rem+5)/10, unitArray[i])
}
