package wget

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
)

const (
	timeWithDHMS = "%d %H %M %S"
	timeWithHMS = "%H %M %S"
)

func toString(s string) string {

	num:=len(s)
	if num <= 3 {
		return s
	}

	n:=num%3
	str:=s[:n]
	for i:=n; i<num; i+=3 {
		str += ","
		str += s[i:i+3]
	}
	return strings.TrimLeft(str, ",")
}

func toStringI(i int) string {

	s := strconv.Itoa(i)
	return toString(s)
}

func toStringI64(i int64) string {

	s := strconv.FormatInt(i, 10)
	return toString(s)
}

func formatTime(t int, format string) string {

	d:=t/86400
	t=t-d*86400
	h:=t/3600
	t=t-h*3600
	m:=t/60
	t=t-m*60

	sd:=""
	if d<10 {
		sd="0"
	}
	sd += strconv.Itoa(d) + "d"

	sh:=""
	if h<10 {
		sh="0"
	}
	sh += strconv.Itoa(h) + "h"

	sm:=""
	if m<10 {
		sm="0"
	}
	sm += strconv.Itoa(m) + "m"

	ss:=""
	if t<10 {
		ss="0"
	}
	ss += strconv.Itoa(t) + "s"

	var ret string
	switch format {
	case timeWithDHMS:
		ret = fmt.Sprintf("%s %s %s %s", sd, sh, sm, ss)
		break;
	case timeWithHMS:
		ret = fmt.Sprintf("%s %s %s", sh, sm, ss)
		break;
	default:
		fmt.Println("not support at present.")
		break;
	}
	return ret
}

func isUrl(url string) bool {

	return strings.Index(url, "http://") == 0 || strings.Index(url, "https://") == 0
}

func formatUrl(url string) string {

	pos := strings.Index(url, "://")
	s := url[pos+3:]
	for i:=0; i<3; i++ {
		s = strings.Replace(s,"//", "/", -1)
	}
	return url[:pos+3]+s
}

func indexUrl(url string, n int) int {

	numSlash := 0
	pos := strings.IndexFunc(url, func(r rune) bool {
		if r == '/' {
			numSlash++
			return numSlash == n
		}
		return false
	})

	return pos
}

func isExist(path string) bool {

	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func createDirectory(dir string) error {

	if !isExist(dir) {
		err := os.MkdirAll(dir, os.ModePerm)
		return err
	}
	return nil
}

func createDirectory2(filePath string) error {

	pos := strings.LastIndex(filePath, "/")
	return createDirectory(filePath[:pos])
}

func getFileName(filePath string) string  {

	pos := strings.LastIndex(filePath, "/")
	return filePath[pos+1:]
}

func getWritePath(outPath, url string, isdir bool) (dir, file string) {

	pos := indexUrl(url, 2)
	dir = path.Join(outPath, url[pos:])
	if isdir {

		pos = strings.LastIndex(dir, "/")
		lenPath := len(dir)
		npos := lenPath
		if pos == lenPath - 1 {
			pos = strings.LastIndex(dir[:lenPath-1], "/")
			npos = lenPath - 1
		}

		file = path.Join(dir, dir[pos:npos])
		file += ".html"
	} else {
		file = dir
		dir = ""
	}
	return
}

func clamp(v, min, max int) int {

	if v <= min {
		return min
	} else if v >= max {
		return max
	} else {
		return v
	}
}