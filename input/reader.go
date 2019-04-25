package input

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Confirm asks for user input and return true on Y or y. Additional strings can be passed that will be printed
// before the confirmation prompt.
func Confirm(notice ...string) bool {
	for n := range notice {
		fmt.Print(notice[n])
	}

	fmt.Print("Please Confirm [Y/n]: ")
	r := bufio.NewReader(os.Stdin)
	res, _ := r.ReadString(byte('\n'))
	res = strings.TrimSuffix(res, "\n")

	if res == "Y" || res == "y" {
		return true
	}

	return false
}

// ConfirmByTyping returns true if the user enters the specified string. Additional strings can be passed that will
// be printed before the confirmation prompt.
func ConfirmByTyping(confirmation string, notice ...string) bool {
	for n := range notice {
		fmt.Print(notice[n])
	}

	fmt.Printf("To confirm please type [%s] without brackets: ", confirmation)
	r := bufio.NewReader(os.Stdin)
	res, _ := r.ReadString(byte('\n'))
	res = strings.TrimSuffix(res, "\n")

	if res == confirmation {
		return true
	}

	return false
}
