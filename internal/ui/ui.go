package ui

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
	colorBold   = "\033[1m"
)

var noColor = os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb"

func paint(color, s string) string {
	if noColor {
		return s
	}
	return color + s + colorReset
}

func Bold(s string) string { return paint(colorBold, s) }

func ClearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func Banner() {
	if noColor {
		fmt.Println(strings.Repeat("=", 54))
		fmt.Println("  OneBuild  -  Flutter cross-platform CI builder")
		fmt.Println("  by A.M.Ghaderi  |  github.com/ghaderi0x")
		fmt.Println(strings.Repeat("=", 54))
		return
	}
	fmt.Println(paint(colorCyan, "┌────────────────────────────────────────────────────┐"))
	fmt.Println(paint(colorCyan, "│") + paint(colorBold+colorPurple, "   OneBuild") + paint(colorGray, "  ·  Flutter cross-platform CI builder") + paint(colorCyan, "  │"))
	fmt.Println(paint(colorCyan, "│") + paint(colorGray, "   by A.M.Ghaderi  ·  github.com/ghaderi0x           ") + paint(colorCyan, "│"))
	fmt.Println(paint(colorCyan, "└────────────────────────────────────────────────────┘"))
}

func Info(format string, a ...interface{}) {
	fmt.Println(paint(colorBlue, "  ›") + " " + fmt.Sprintf(format, a...))
}

func Success(format string, a ...interface{}) {
	fmt.Println(paint(colorGreen, "  ✔") + " " + fmt.Sprintf(format, a...))
}

func Warn(format string, a ...interface{}) {
	fmt.Println(paint(colorYellow, "  ⚠") + " " + fmt.Sprintf(format, a...))
}

func Error(format string, a ...interface{}) {
	fmt.Println(paint(colorRed, "  ✖") + " " + fmt.Sprintf(format, a...))
}

func Step(format string, a ...interface{}) {
	fmt.Println()
	fmt.Println(paint(colorBold+colorPurple, "▶ ") + paint(colorBold, fmt.Sprintf(format, a...)))
}

func Divider() {
	fmt.Println(paint(colorGray, strings.Repeat("─", 54)))
}

type Spinner struct {
	mu      sync.Mutex
	message string
	stopCh  chan struct{}
	doneCh  chan struct{}
	active  bool
}

func NewSpinner() *Spinner {
	return &Spinner{}
}

func (s *Spinner) Start(message string) {
	s.mu.Lock()
	if s.active {
		s.message = message
		s.mu.Unlock()
		return
	}
	s.message = message
	s.stopCh = make(chan struct{})
	s.doneCh = make(chan struct{})
	s.active = true
	s.mu.Unlock()

	go func() {
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		i := 0
		ticker := time.NewTicker(90 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-s.stopCh:
				fmt.Print("\r\033[K")
				close(s.doneCh)
				return
			case <-ticker.C:
				s.mu.Lock()
				msg := s.message
				s.mu.Unlock()
				frame := frames[i%len(frames)]
				if noColor {
					fmt.Printf("\r  %s %s", frame, msg)
				} else {
					fmt.Printf("\r  %s%s%s %s", colorCyan, frame, colorReset, msg)
				}
				i++
			}
		}
	}()
}

func (s *Spinner) UpdateMessage(message string) {
	s.mu.Lock()
	s.message = message
	s.mu.Unlock()
}

func (s *Spinner) Stop() {
	s.mu.Lock()
	if !s.active {
		s.mu.Unlock()
		return
	}
	s.active = false
	ch := s.stopCh
	done := s.doneCh
	s.mu.Unlock()
	close(ch)
	<-done
}

var stdinReader = bufio.NewReader(os.Stdin)

func AskString(label, defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf("  %s %s: ", paint(colorCyan, "?"), fmt.Sprintf("%s [%s]", label, defaultValue))
	} else {
		fmt.Printf("  %s %s: ", paint(colorCyan, "?"), label)
	}
	line, _ := stdinReader.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return defaultValue
	}
	return line
}

func AskSecret(label string) string {
	fmt.Printf("  %s %s: ", paint(colorCyan, "?"), label)
	value := readSecretLine()
	fmt.Println()
	return strings.TrimSpace(value)
}

func readSecretLine() string {
	restoreEcho := func() {}
	if runtime.GOOS != "windows" {
		cmd := exec.Command("stty", "-echo")
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err == nil {
			restoreEcho = func() {
				restore := exec.Command("stty", "echo")
				restore.Stdin = os.Stdin
				restore.Run()
			}
		}
	}
	defer restoreEcho()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	defer func() {
		signal.Stop(sigCh)
		close(sigCh)
	}()
	go func() {
		if _, ok := <-sigCh; ok {
			restoreEcho()
			fmt.Println()
			os.Exit(130)
		}
	}()

	line, _ := stdinReader.ReadString('\n')
	return line
}

func AskYesNo(label string, defaultYes bool) bool {
	suffix := "Y/n"
	if !defaultYes {
		suffix = "y/N"
	}
	fmt.Printf("  %s %s (%s): ", paint(colorCyan, "?"), label, suffix)
	line, _ := stdinReader.ReadString('\n')
	line = strings.ToLower(strings.TrimSpace(line))
	if line == "" {
		return defaultYes
	}
	return line == "y" || line == "yes"
}

func AskChoice(label string, options []string) int {
	fmt.Printf("  %s %s\n", paint(colorCyan, "?"), label)
	for i, opt := range options {
		fmt.Printf("      %s) %s\n", paint(colorBold, fmt.Sprintf("%d", i+1)), opt)
	}
	for {
		fmt.Printf("  %s Enter number: ", paint(colorCyan, ">"))
		line, _ := stdinReader.ReadString('\n')
		line = strings.TrimSpace(line)
		var n int
		if _, err := fmt.Sscanf(line, "%d", &n); err == nil && n >= 1 && n <= len(options) {
			return n - 1
		}
		Warn("Invalid choice, try again.")
	}
}

func AskMultiChoice(label string, options []string) []int {
	fmt.Printf("  %s %s (comma separated numbers, e.g. 1,3)\n", paint(colorCyan, "?"), label)
	for i, opt := range options {
		fmt.Printf("      %s) %s\n", paint(colorBold, fmt.Sprintf("%d", i+1)), opt)
	}
	for {
		fmt.Printf("  %s Enter numbers: ", paint(colorCyan, ">"))
		line, _ := stdinReader.ReadString('\n')
		line = strings.TrimSpace(line)
		parts := strings.Split(line, ",")
		seen := map[int]bool{}
		var result []int
		valid := len(parts) > 0
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			var n int
			if _, err := fmt.Sscanf(p, "%d", &n); err != nil || n < 1 || n > len(options) {
				valid = false
				break
			}
			if !seen[n] {
				seen[n] = true
				result = append(result, n-1)
			}
		}
		if valid && len(result) > 0 {
			return result
		}
		Warn("Invalid choice, try again.")
	}
}
