package internal

import (
	"bytes"
	"fmt"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"voidedtech.com/stock"
)

const (
	getCommand  = "get:"
	setCommand  = "set:"
	respCommand = "res:"
)

var (
	credential []byte
	lock       = &sync.Mutex{}
	stored     time.Time
)

func readConn(conn net.Conn) (string, error) {
	buf := make([]byte, 512)
	if _, err := conn.Read(buf); err != nil {
		return "", err
	}
	b := bytes.Trim(buf, "\x00")
	return strings.TrimSpace(string(b)), nil
}

func purge(duration time.Duration) {
	for {
		lock.Lock()
		if credential != nil {
			now := time.Now().Add(duration)
			if stored.Before(now) {
				credential = nil
				stored = time.Now()
			}
		}
		lock.Unlock()
		time.Sleep(5 * time.Second)
	}
}

// SocketHandler handles the daemon socket for lockbox key resolution.
func SocketHandler(isHost bool) error {
	path := os.Getenv("LOCKBOX_SOCKET")
	if path == "" {
		h := os.Getenv("HOME")
		if h == "" {
			return NewLockboxError("unable to get HOME")
		}
		path = filepath.Join(h, ".lb", "lockbox.sock")
	}
	if isHost {
		caching := 1440
		if keep := os.Getenv("LOCKBOX_CCACHE"); keep != "" {
			i, err := strconv.Atoi(keep)
			if err != nil {
				return err
			}
			caching = i
		}
		if caching != 0 {
			if caching > 0 {
				caching *= -1
			}
			keepFor := time.Duration(caching) * time.Minute
			go purge(keepFor)
		}
		dir := filepath.Dir(path)
		if !stock.PathExists(dir) {
			if err := os.MkdirAll(dir, 0700); err != nil {
				return err
			}
		}
		stats, err := os.Stat(dir)
		if err != nil {
			return err
		}
		if stats.Mode() != fs.ModeDir|0700 {
			return NewLockboxError("invalid permissions on lb socket directory, too open")
		}
		if stock.PathExists(path) {
			if err := os.Remove(path); err != nil {
				return err
			}
		}
		l, err := net.Listen("unix", path)
		if err != nil {
			return err
		}
		defer l.Close()
		for {
			conn, err := l.Accept()
			if err != nil {
				stock.LogError("unable to accept connection", err)
				continue
			}
			cmd, err := readConn(conn)
			if err != nil {
				stock.LogError("failed to read command", err)
				conn.Close()
				continue
			}
			lock.Lock()
			if strings.HasPrefix(cmd, getCommand) {
				write := []byte(respCommand)
				if credential != nil {
					write = append(write, credential...)
				}
				_, err := conn.Write(write)
				if err != nil {
					stock.LogError("failed to write credential to connection", err)
				}
			} else {
				if strings.HasPrefix(cmd, setCommand) {
					text := strings.Replace(cmd, setCommand, "", 1)
					credential = []byte(text)
					stored = time.Now()
					if _, err := conn.Write([]byte(respCommand)); err != nil {
						stock.LogError("failed to write empty set response", err)
					}
				} else {
					stock.LogError("unknown command", nil)
				}
			}
			lock.Unlock()
			conn.Close()
		}
	}

	c, err := net.Dial("unix", path)
	if err != nil {
		return err
	}
	_, err = c.Write([]byte(getCommand))
	if err != nil {
		c.Close()
		return err
	}
	data, err := readConn(c)
	c.Close()
	if err != nil {
		return err
	}
	if data == respCommand {
		termEcho(false)
		input, err := Stdin(true)
		termEcho(true)
		if err != nil {
			return err
		}
		setting := []byte(setCommand)
		setting = append(setting, input...)
		c, err := net.Dial("unix", path)
		if err != nil {
			return err
		}
		if _, err := c.Write(setting); err != nil {
			return err
		}
		data = input
	} else {
		data = strings.Replace(data, respCommand, "", 1)
	}
	fmt.Println(data)
	return nil
}
