package pid

import (
	"encoding/json"
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
var ServerInfoFile string

const appName = "mock-todo-server"

type ServerInfo struct {
	PID  int `json:"pid"`
	Port int `json:"port"`
}

func init() {
	if runtime.GOOS != "windows" {
		pidPath, err := xdg.RuntimeFile(filepath.Join(appName, appName+".pid"))
		if err == nil {
			if err := os.MkdirAll(filepath.Dir(pidPath), 0700); err != nil {
				log.Fatal("failed to create pid directory:", err)
			}
			PidFile = pidPath
			ServerInfoFile = filepath.Join(filepath.Dir(pidPath), appName+".json")
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
	ServerInfoFile = filepath.Join(appDir, appName+".json")
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

func CreateServerInfoFile(pid, port int) error {
	serverInfo := ServerInfo{
		PID:  pid,
		Port: port,
	}

	data, err := json.MarshalIndent(serverInfo, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(ServerInfoFile, data, 0600); err != nil {
		return err
	}

	log.Printf("server info file created with PID: %d, Port: %d", pid, port)
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

func GetServerInfo() (*ServerInfo, error) {
	data, err := os.ReadFile(ServerInfoFile)
	if err != nil {
		return nil, err
	}

	var serverInfo ServerInfo
	if err := json.Unmarshal(data, &serverInfo); err != nil {
		return nil, err
	}

	return &serverInfo, nil
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

func RemoveServerInfoFile() error {
	if err := os.Remove(ServerInfoFile); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
