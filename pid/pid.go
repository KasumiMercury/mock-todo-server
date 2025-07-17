package pid

import (
	"github.com/adrg/xdg"
	"log"
	"os"
	"path/filepath"
	"runtime"
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
			log.Println("creating pid file at", pidPath)
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
	log.Println("creating pid file at", PidFile)
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

	_, err = file.WriteString(string(rune(pid)))
	if err != nil {
		return err
	}

	log.Println("pid file created with PID:", pid)
	return nil
}
