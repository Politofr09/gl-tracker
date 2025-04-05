package utils

import (
	"fmt"
)

func PrintFancy(source string, color string, a ... any) {
	fmt.Print(color, "[", source, "]\033[0m ")
	fmt.Print(a...)
	fmt.Print("\n")
}