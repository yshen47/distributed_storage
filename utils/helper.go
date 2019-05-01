package utils

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
)

func SetupLog(name string) *os.File {
	path := Concatenate("logs/", name, ".txt")
	_ = os.Remove(path)

	f, err := os.OpenFile(path, os.O_RDWR | os.O_CREATE, 0666)
	CheckError(err, true)

	log.SetOutput(f)
	//log.SetPrefix(name + " ")
	//log.Print("#########################################")
	//log.Println("Start coordinator...")
	return f
}

func Concatenate(elem ...interface{}) string {
	var ipAddress []string
	for _, e := range elem {
		switch v := e.(type) {
		case string:
			ipAddress = append(ipAddress, v)
		case int:
			t := strconv.Itoa(v)
			ipAddress = append(ipAddress, t)
		case float64:
			t := fmt.Sprintf("%f",v)
			ipAddress = append(ipAddress,t)
		default:
			fmt.Printf("unexpected type %T", v)
		}
	}

	return strings.Join(ipAddress, "")
}


func CheckError(err error, exit bool) {
	if err != nil {
		_, fn, line, _ := runtime.Caller(1)
		fmt.Printf("[error] %s:%d %v", fn, line, err)
		if exit {
			os.Exit(7)
		}
	}
}

