package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/term"
)

// ====== Nordic Color Palette ======

const (
	Nord0  = "\033[38;5;235m"
	Nord1  = "\033[38;5;236m"
	Nord2  = "\033[38;5;237m"
	Nord3  = "\033[38;5;239m"
	Nord4  = "\033[38;5;245m"
	Nord5  = "\033[38;5;250m"
	Nord6  = "\033[38;5;254m"
	Nord7  = "\033[38;5;153m"
	Nord8  = "\033[38;5;81m"
	Nord9  = "\033[38;5;110m"
	Nord10 = "\033[38;5;109m"
	Nord11 = "\033[38;5;167m"
	Nord12 = "\033[38;5;142m"
	Nord13 = "\033[38;5;180m"
	Nord14 = "\033[38;5;108m"
	Nord15 = "\033[38;5;182m"
	Reset  = "\033[0m"
)

// ====== Utility functions ======

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func colorText(text string, colorCode string) string {
	return fmt.Sprintf("%s%s%s", colorCode, text, Reset)
}

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

func centerPrint(text string) {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = 80 // fallback
	}
	visibleText := stripANSI(text)
	textLen := len([]rune(visibleText))
	padding := (width - textLen) / 2
	if padding > 0 {
		fmt.Print(strings.Repeat(" ", padding))
	}
	fmt.Println(text)
}

func showHomeScreen() {
	clearScreen()
	centerPrint(colorText("╔══════════════════════════════════════╗", Nord8))
	centerPrint(colorText("║           Welcome to Tchat           ║", Nord7))
	centerPrint(colorText("╚══════════════════════════════════════╝", Nord8))
	fmt.Println()
	centerPrint(colorText("1) Create a Room", Nord4))
	centerPrint(colorText("2) Join a Room", Nord4))
	fmt.Println()
	centerPrint(colorText("Enter choice (1 or 2):", Nord3))
	fmt.Print("> ")
}

// ====== Main Client ======

func main() {
	conn, err := net.Dial("tcp", "localhost:9001")
	if err != nil {
		fmt.Println("Connection error:", err)
		return
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)

	clearScreen()
	centerPrint(colorText("Welcome to Tchat!", Nord7))
	fmt.Println()

	// Ask for Display Name
	centerPrint(colorText("Enter your display name:", Nord4))
	fmt.Print("> ")
	displayName, _ := reader.ReadString('\n')
	displayName = strings.TrimSpace(displayName)

	// Show home screen and get choice
	showHomeScreen()
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	for choice != "1" && choice != "2" {
		centerPrint("Invalid choice. Please enter 1 or 2:")
		fmt.Print("> ")
		choice, _ = reader.ReadString('\n')
		choice = strings.TrimSpace(choice)
	}

	clientID, err := LoadOrCreateClientID()
	if err != nil {
		fmt.Println("Failed to load client ID:", err)
		return
	}

	var roomData string
	if choice == "1" {
		centerPrint(colorText("Enter room capacity (MAX: 20):", Nord3))
		fmt.Print("> ")

		var capInt int
		for {
			roomData, _ = reader.ReadString('\n')
			roomData = strings.TrimSpace(roomData)
			capInt, err = strconv.Atoi(roomData)
			if err == nil && capInt >= 1 && capInt <= 20 {
				break
			}
			centerPrint(colorText("Invalid room capacity. Must be a number from 1 to 20.", Nord11))
			fmt.Print("> ")
		}
	} else if choice == "2" {
		centerPrint(colorText("Enter room ID to join:", Nord3))
		fmt.Print("> ")
		roomData, _ = reader.ReadString('\n')
		roomData = strings.TrimSpace(roomData)
	}

	finalMessage := fmt.Sprintf("%d|%s|%s|%s", clientID, displayName, choice, roomData)
	_, err = fmt.Fprintln(conn, finalMessage)
	if err != nil {
		fmt.Println("Send error:", err)
		return
	}

	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()

	for {
		userInput, _ := reader.ReadString('\n')
		userInput = strings.TrimSpace(userInput)
		_, err = fmt.Fprintln(conn, userInput)
		if err != nil {
			fmt.Println("Send error:", err)
			break
		}
	}
}
