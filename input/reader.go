package input

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type InputReader struct {
}

func New() *InputReader {
	return &InputReader{}
}

// Confirm asks for user input and return true on Y or y. Additional strings can be passed that will be printed
// before the confirmation prompt.
func (i *InputReader) Confirm(notice ...string) bool {
	for n := range notice {
		fmt.Print(notice[n])
	}

	fmt.Print("Please Confirm [Y/n]: ")
	r := bufio.NewReader(os.Stdin)
	res, _ := r.ReadString(byte('\n'))
	res = strings.TrimSuffix(res, "\n")

	if res != "Y" && res != "y" {
		return false
	}

	return true
}
