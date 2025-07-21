package interactive

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func displayOneLiner(commands []string, flags []string) {
	fmt.Printf("\n--- Command line equivalent ---\n")

	executableName := filepath.Base(os.Args[0])

	var cmdParts []string
	cmdParts = append(cmdParts, executableName)
	cmdParts = append(cmdParts, commands...)
	cmdParts = append(cmdParts, flags...)

	fmt.Printf("%s\n\n", strings.Join(cmdParts, " "))
}
