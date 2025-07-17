package pid

import (
	"fmt"
	"github.com/adrg/xdg"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

var PidFile string

const appName = "mock-todo-server"

func init() {
	if runtime.GOOS != "windows" {
		pidPath, err := xdg.RuntimeFile(filepath.Join(appName, appName+".pid"))
		if err == nil {
			if err := os.MkdirAll(filepath.Dir(pidPath), 0700); err != nil {
				log.Fatal("failed to create pid directory:", err)
			}
			PidFile = pidPath
			log.Println("pid file path:", PidFile)
			return
		}
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Fatal(err)
	}

	appDir := filepath.Join(cacheDir, appName)
	if err := os.MkdirAll(appDir, 0700); err != nil {
		log.Fatal("failed to create app directory:", err)
	}

	PidFile = filepath.Join(appDir, appName+".pid")
	log.Println("pid file path:", PidFile)
}

func CheckRunning() bool {
	if _, err := os.Stat(PidFile); err == nil {
		log.Println("pid file exists, server is already running")
		return true
	}

	return false
}

func CreatePidFile(pid int) error {
	file, err := os.Create(PidFile)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("%d", pid))
	if err != nil {
		return err
	}

	log.Println("pid file created with PID:", pid)
	return nil
}

func GetPid() (int, error) {
	data, err := os.ReadFile(PidFile)
	if err != nil {
		return 0, err
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, err
	}

	return pid, nil
}

func StopByPid() error {
	pid, err := GetPid()
	if err != nil {
		return fmt.Errorf("failed to read PID: %w", err)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}

	if err := process.Signal(os.Interrupt); err != nil {
		return fmt.Errorf("failed to send signal: %w", err)
	}

	log.Printf("sent interrupt signal to process %d", pid)
	return nil
}
