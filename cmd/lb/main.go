package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"voidedtech.com/lockbox/internal"
)

func getEntry(store string, args []string, idx int) string {
	if len(args) != idx+1 {
		internal.Die("invalid entry given", fmt.Errorf("specific entry required"))
	}
	return filepath.Join(store, args[idx]) + internal.Extension
}

func termEcho(on bool) {
	// Common settings and variables for both stty calls.
	attrs := syscall.ProcAttr{
		Dir:   "",
		Env:   []string{},
		Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd()},
		Sys:   nil}
	var ws syscall.WaitStatus
	cmd := "echo"
	if !on {
		cmd = "-echo"
	}

	// Enable/disable echoing.
	pid, err := syscall.ForkExec(
		"/bin/stty",
		[]string{"stty", cmd},
		&attrs)
	if err != nil {
		panic(err)
	}

	// Wait for the stty process to complete.
	_, err = syscall.Wait4(pid, &ws, 0, nil)
	if err != nil {
		panic(err)
	}
}

func readInput() (string, error) {
	termEcho(false)
	defer func() {
		termEcho(true)
	}()
	fmt.Printf("please enter password: ")
	first, err := stdin(true)
	if err != nil {
		return "", err
	}
	fmt.Printf("\nplease re-enter password: ")
	second, err := stdin(true)
	if err != nil {
		return "", err
	}
	if first != second {
		return "", fmt.Errorf("passwords do NOT match")
	}
	return first, nil
}

func pipeTo(command, value string, wait bool, args ...string) {
	cmd := exec.Command(command, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		internal.Die("unable to get stdin pipe", err)
	}

	go func() {
		defer stdin.Close()
		if _, err := stdin.Write([]byte(value)); err != nil {
			fmt.Printf("failed writing to stdin: %v\n", err)
		}
	}()
	var ran error
	if wait {
		ran = cmd.Run()
	} else {
		ran = cmd.Start()
	}
	if ran != nil {
		internal.Die("failed to run command", ran)
	}
}

func clipboard(value string) {
	pipeTo("pbcopy", value, true)
}

func main() {
	args := os.Args
	if len(args) < 2 {
		internal.Die("missing arguments", fmt.Errorf("requires subcommand"))
	}
	command := args[1]
	store := internal.GetStore()
	switch command {
	case "ls", "list":
		files, err := internal.Find(store, true)
		if err != nil {
			internal.Die("unable to list files", err)
		}
		for _, f := range files {
			fmt.Println(f)
		}
	case "insert":
		multi := false
		idx := 2
		switch len(args) {
		case 2:
			internal.Die("insert missing required arguments", fmt.Errorf("entry required"))
		case 3:
		case 4:
			multi = args[2] == "-m"
			if !multi {
				internal.Die("multi-line insert must be after 'insert'", fmt.Errorf("invalid command"))
			}
			idx = 3
		default:
			internal.Die("too many arguments", fmt.Errorf("insert can only perform one operation"))
		}
		isPipe := isInputFromPipe()
		entry := getEntry(store, args, idx)
		if internal.PathExists(entry) {
			if !isPipe {
				if !confirm("overwrite existing") {
					return
				}
			}
		} else {
			dir := filepath.Dir(entry)
			if !internal.PathExists(dir) {
				if err := os.MkdirAll(dir, 0755); err != nil {
					internal.Die("failed to create directory structure", err)
				}
			}
		}
		password := ""
		if !multi && !isPipe {
			input, err := readInput()
			if err != nil {
				internal.Die("password input failed", err)
			}
			password = input
		} else {
			input, err := stdin(false)
			if err != nil {
				internal.Die("failed to read stdin", err)
			}
			password = input
		}
		if password == "" {
			internal.Die("empty password provided", fmt.Errorf("password can NOT be empty"))
		}
		l, err := internal.NewLockbox("", "", entry)
		if err != nil {
			internal.Die("unable to make lockbox model instance", err)
		}
		if err := l.Encrypt([]byte(password)); err != nil {
			internal.Die("failed to save password", err)
		}
		fmt.Println("")
	case "rm":
		entry := getEntry(store, args, 2)
		if !internal.PathExists(entry) {
			internal.Die("does not exists", fmt.Errorf("can not delete unknown entry"))
		}
		if confirm("remove entry") {
			os.Remove(entry)
		}
	case "show", "-c", "clip":
		entry := getEntry(store, args, 2)
		if !internal.PathExists(entry) {
			internal.Die("invalid entry", fmt.Errorf("entry not found"))
		}
		l, err := internal.NewLockbox("", "", entry)
		if err != nil {
			internal.Die("unable to make lockbox model instance", err)
		}
		decrypt, err := l.Decrypt()
		if err != nil {
			internal.Die("unable to decrypt", err)
		}
		value := strings.TrimSpace(string(decrypt))
		if command == "show" {
			fmt.Println(value)
			return
		}
		clipboard(value)
		fmt.Println("clipboard will clear in 45 seconds")
		pipeTo("lb", value, false, "clear")
	case "clear":
		idx := 0
		val, err := stdin(false)
		if err != nil {
			internal.Die("unable to read value to clear", err)
		}
		val = strings.TrimSpace(val)
		for idx < 45 {
			idx++
			time.Sleep(1 * time.Second)
			out, err := exec.Command("pbpaste").Output()
			if err != nil {
				continue
			}
			fmt.Println(string(out))
			fmt.Println(val)
			if strings.TrimSpace(string(out)) != val {
				return
			}
		}
		clipboard("")
	default:
		tryCommand := fmt.Sprintf("lb-%s", command)
		c := exec.Command(tryCommand, args[2:]...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			internal.Die("bad command", err)
		}
	}
}

func stdin(one bool) (string, error) {
	b, err := internal.Stdin(one)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

func confirm(prompt string) bool {
	fmt.Printf("%s? (y/N) ", prompt)
	resp, err := stdin(true)
	if err != nil {
		internal.Die("failed to get response", err)
	}
	return resp == "Y" || resp == "y"
}

func isInputFromPipe() bool {
	fileInfo, _ := os.Stdin.Stat()
	return fileInfo.Mode()&os.ModeCharDevice == 0
}
