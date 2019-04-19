package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"math/rand"
	"net"
	"os"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func SetupLog(name string) *os.File {
	path := Concatenate("logs/", name, ".txt")
	_ = os.Remove(path)

	f, err := os.OpenFile(path, os.O_RDWR | os.O_CREATE, 0666)
	CheckError(err)

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


func CheckError(err error) {
	if err != nil {
		_, fn, line, _ := runtime.Caller(1)
		fmt.Printf("[error] %s:%d %v", fn, line, err)
		os.Exit(7)
	}
}

func Shuffle(vals []int) []int {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	ret := make([]int, len(vals))
	perm := r.Perm(len(vals))
	for i, randIndex := range perm {
		ret[i] = vals[randIndex]
	}
	return ret
}

func Arange(start, stop, step int) []int {
	N := (stop - start) / step
	rnge := make([]int, N, N)
	i := 0
	for x := start; x < stop; x += step {
		rnge[i] = x;
		i += 1
	}
	return rnge
}

func GetCurrentIP(debug bool, port int) string {
	if !debug {
		addrs, err := net.InterfaceAddrs()
		if err != nil {
			os.Stderr.WriteString("Oops: " + err.Error() + "\n")
			os.Exit(8)
		}

		for _, a := range addrs {
			if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return Concatenate(ipnet.IP.String(), ":", port)
				}
			}
		}
		return ""
	} else {
		return Concatenate("127.0.0.1:", port)
	}
}

func StringAddrToIntArr(addr string) []byte {
	rawArr := strings.Split(addr, ":")
	addrArr := strings.Split(rawArr[0], ".")
	var res = [] byte {}

	for _, i := range addrArr {
		j, err := strconv.Atoi(i)
		if err != nil {
			panic(err)
		}
		res = append(res, byte(j))
	}

	return res
}

func GetSHA256(input []byte) string {
	hash := sha256.New()
	hash.Write(input)
	md := hash.Sum(nil)
	mdStr := hex.EncodeToString(md)
	return mdStr
}
