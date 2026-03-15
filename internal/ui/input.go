package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var reader = bufio.NewReader(os.Stdin)

// ReadString reads a string from stdin with a prompt
func ReadString(prompt string) string {
	fmt.Print(Cyan + "  " + prompt + ": " + Reset)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// ReadInt reads an integer from stdin with range validation
func ReadInt(prompt string, min, max int) int {
	for {
		str := ReadString(fmt.Sprintf("%s [%d-%d]", prompt, min, max))
		val, err := strconv.Atoi(str)
		if err != nil || val < min || val > max {
			ShowError(fmt.Sprintf("Ingresa un número entre %d y %d", min, max))
			continue
		}
		return val
	}
}

// ReadOption reads a menu option (single character, no range enforcement)
func ReadOption() string {
	fmt.Print(Cyan + "\n  Selecciona una opción: " + Reset)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// Confirm asks a yes/no question
func Confirm(prompt string) bool {
	fmt.Print(Yellow + "  " + prompt + " (s/n): " + Reset)
	input, _ := reader.ReadString('\n')
	answer := strings.ToLower(strings.TrimSpace(input))
	return answer == "s" || answer == "si" || answer == "sí" || answer == "y" || answer == "yes"
}

// Pause waits for Enter key
func Pause() {
	fmt.Print(Dim + "\n  Presiona Enter para continuar..." + Reset)
	reader.ReadString('\n')
}
