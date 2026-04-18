package utils

import (
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

func UUID() string {
	code := uuid.New().String()
	code = strings.Replace(code, "-", "", -1)
	return code
}

const (
	keyChars   = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	keyNumbers = "0123456789"
)

func GenerateKey() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	key := make([]byte, 48)
	for i := 0; i < 16; i++ {
		key[i] = keyChars[r.Intn(len(keyChars))]
	}

	id := UUID()
	for i := 0; i < 32; i++ {
		c := id[i]
		if i%2 == 0 && c >= 'a' && c <= 'z' {
			c = c - 'a' + 'A'
		}

		key[i+16] = c
	}

	return string(key)
}

func GetRandomString(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	key := make([]byte, length)
	for i := 0; i < length; i++ {
		key[i] = keyChars[r.Intn(len(keyChars))]
	}

	return string(key)
}

func GetRandomNumberString(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	key := make([]byte, length)
	for i := 0; i < length; i++ {
		key[i] = keyNumbers[r.Intn(len(keyNumbers))]
	}

	return string(key)
}

// RandRange returns a random number between min and max (max is not included)
func RandRange(min, max int) int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	return min + r.Intn(max-min)
}

func OpenBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	}
	if err != nil {
		log.Println(err)
	}
}

func GetIp() (ip string) {
	ips, err := net.InterfaceAddrs()
	if err != nil {
		log.Println(err)
		return ip
	}

	for _, a := range ips {
		if ipNet, ok := a.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ip = ipNet.IP.String()
				if strings.HasPrefix(ip, "10") {
					return
				}
				if strings.HasPrefix(ip, "172") {
					return
				}
				if strings.HasPrefix(ip, "192.168") {
					return
				}
				ip = ""
			}
		}
	}

	return
}

const (
	sizeKB = 1024
	sizeMB = sizeKB * 1024
	sizeGB = sizeMB * 1024
)

func Bytes2Size(num int64) string {
	numStr := ""
	unit := "B"
	if num/int64(sizeGB) > 1 {
		numStr = fmt.Sprintf("%.2f", float64(num)/float64(sizeGB))
		unit = "GB"
	} else if num/int64(sizeMB) > 1 {
		numStr = fmt.Sprintf("%d", int(float64(num)/float64(sizeMB)))
		unit = "MB"
	} else if num/int64(sizeKB) > 1 {
		numStr = fmt.Sprintf("%d", int(float64(num)/float64(sizeKB)))
		unit = "KB"
	} else {
		numStr = fmt.Sprintf("%d", num)
	}

	return numStr + " " + unit
}

func Interface2String(inter interface{}) string {
	switch inter := inter.(type) {
	case string:
		return inter
	case int:
		return fmt.Sprintf("%d", inter)
	case float64:
		return fmt.Sprintf("%f", inter)
	}

	return "Not Implemented"
}

func UnescapeHTML(x string) interface{} {
	return template.HTML(x)
}

func IntMax(a int, b int) int {
	if a >= b {
		return a
	}

	return b
}

func Max(a int, b int) int {
	if a >= b {
		return a
	}

	return b
}

func AssignOrDefault(value string, defaultValue string) string {
	if len(value) != 0 {
		return value
	}

	return defaultValue
}

func MessageWithRequestId(message string, id string) string {
	return fmt.Sprintf("%s (request id: %s)", message, id)
}

func String2Int(str string) int {
	num, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return num
}

func Float64PtrMax(p *float64, maxValue float64) *float64 {
	if p == nil {
		return nil
	}
	if *p > maxValue {
		return &maxValue
	}
	return p
}

func Float64PtrMin(p *float64, minValue float64) *float64 {
	if p == nil {
		return nil
	}

	if *p < minValue {
		return &minValue
	}
	return p
}
