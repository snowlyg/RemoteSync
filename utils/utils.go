package utils

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func CWD() string {
	path, err := os.Executable()
	if err != nil {
		return ""
	}
	return filepath.Dir(path)
}

func EXEName() string {
	path, err := os.Executable()
	if err != nil {
		return ""
	}
	return strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
}

func LogDir() string {
	dir := filepath.Join(CWD(), "logs")
	EnsureDir(dir)
	return dir
}

var FlagVarDBFile string

func ConfigFile() string {
	if FlagVarDBFile != "" {
		return FlagVarDBFile
	}
	if Exist(DBFileDev()) {
		return DBFileDev()
	}
	return filepath.Join(CWD(), "config.yaml")
}

func DBFile() string {
	if FlagVarDBFile != "" {
		return FlagVarDBFile
	}
	if Exist(DBFileDev()) {
		return DBFileDev()
	}
	return filepath.Join(CWD(), strings.ToLower(EXEName()+".db"))
}

func DBFileDev() string {
	return filepath.Join(CWD(), strings.ToLower(EXEName())+".dev.db")
}

func EnsureDir(dir string) (err error) {
	if _, err = os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return
		}
	}
	return
}

func Exist(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func IsPortInUse(host string, port int64) error {
	if conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, fmt.Sprintf("%d", port)), 3*time.Second); err == nil {
		conn.Close()
		return nil
	} else {
		return err
	}

}
