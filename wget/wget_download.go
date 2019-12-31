package wget

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

const (
	threadNum = 100
	packMinSize = threadNum * 1024
)

type downloadInfo struct {

	url string
	timeBegin int64
	timeEnd int64
	err error

	acceptRange bool
	contentLength int64
	contentType string

	size int64
}

type download struct {

	gg *gowget
	aoi aoi

	client *http.Client

	dlList []*downloadInfo
	bufPool *sync.Pool

	stats *downloadStats
}

func newDownload(gg *gowget, aoi aoi) *download {

	return &download{
		gg:gg,
		aoi:aoi,
		stats:newDownloadStats(),
		client:&http.Client{Timeout:time.Second * 180},
		bufPool:&sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(make([]byte, 0, packMinSize))
			},
		},
	}
}

func (dl *download) parseDownloadInfo()  {

	dl.parseHead(dl.gg.url, true)
}

func (dl *download) parseHead(url string, recursive bool)  {

	cliRequest, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		gLog.logErr("failed to new request.[err:%s, url:%s]", err.Error(), url)
		return
	}

	cliResponse, err := dl.client.Do(cliRequest)
	if err != nil {
		gLog.logErr("failed to download head-info.[err:%s, url:%s]", err.Error(), url)
		return
	}
	defer cliResponse.Body.Close()

	cType := cliResponse.Header.Get("Content-Type")
	if recursive && strings.Contains(cType, "html") {

		htmlRequest, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			gLog.logErr("failed to new request.[err:%s, url:%s]", err.Error(), url)
			return
		}

		htmlResponse, err := dl.client.Do(htmlRequest)
		if err != nil {
			gLog.logErr("failed to download html.[err:%s, url:%s]", err.Error(), url)
			return
		}
		defer htmlResponse.Body.Close()

		body, err := ioutil.ReadAll(htmlResponse.Body)
		if err != nil {
			gLog.logErr("failed to read html.[err:%s, url:%s]", err.Error(), url)
		} else {
			dl.writeHtml(url, string(body))
			dl.parseHtml(string(body), dl.gg.flagRecursive)
		}
	} else {

		rangeBytes := cliResponse.Header.Get("Accept-Ranges")
		acceptRange := false
		if strings.Compare(rangeBytes, "bytes") == 0 {
			acceptRange = true
		}
		dl.addDownloadInfo(url, acceptRange, cliResponse.ContentLength, cType)
	}
}

func (dl *download) writeHtml(url, body string)  {

	dir, filePath := dl.getWritePath(url, true)
	createDirectory(dir)

	file, err := os.Create(filePath)
	if err != nil {
		gLog.logErr("failed to create html.[err:%s, url:%s]", err.Error(), url)
		return
	}
	defer file.Close()

	file.WriteString(body)
}

func (dl *download) parseHtml(body string, recursive bool)  {

	result := dl.aoi.getInterest(body)
	num:=len(result)
	for i:=0; i<num; i++ {

		dl.parseHead(result[i], recursive)
	}
}

func (dl *download) addDownloadInfo(url string, acceptRange bool, contentLength int64, contentType string) *downloadInfo {

	dlInfo := &downloadInfo{
		url:url,
		acceptRange:acceptRange,
		contentLength:contentLength,
		contentType:contentType,
		timeBegin:0,
		timeEnd:0,
		err:nil,
		size:0,
	}
	dl.dlList = append(dl.dlList, dlInfo)

	return dlInfo
}

func (dl *download) getWritePath(url string, isdir bool) (dir, file string) {

	pos := indexUrl(url, 2)
	dir = path.Join(dl.gg.flagOutpath, url[pos:])
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

func (dl *download) verbose(vb bool)  {

	if dl.gg.flagVerbose {

		for i:=0; i<len(dl.dlList); i++ {

			strErr := "nil"
			if dl.dlList[i].err != nil {
				strErr = dl.dlList[i].err.Error()
			}
			gLog.logInfo("\ndownload:%s, error:%s", dl.dlList[i].url, strErr)
		}
	}
}

func (dl *download) download()  {

	num := len(dl.dlList)
	for i:=0; i<num; i++ {

		var chunkSize int64
		useThread := true
		info := dl.dlList[i]
		if info.contentLength <= packMinSize {
			chunkSize = info.contentLength
			useThread = false
		} else {
			chunkSize = info.contentLength / threadNum
		}

		normalUrl, err := url.PathUnescape(info.url)
		if err != nil {
			info.err = err
			continue
		}

		_, filePath := dl.getWritePath(normalUrl, false)
		createDirectory2(filePath)
		file, err := os.Create(filePath)
		if err != nil {
			info.err = err
			continue
		}
		defer file.Close()

		info.timeBegin = time.Now().Unix()
		if info.acceptRange && useThread {

			wg := &sync.WaitGroup{}
			fileLock := &sync.Mutex{}

			wg.Add(threadNum)
			for j:=0; j<threadNum; j++ {

				start := int64(j) * chunkSize
				end := start + chunkSize
				if j == threadNum - 1 {
					end = info.contentLength
				}
				go dl.downloadChunk(info, file, start, end-1, wg, fileLock)
			}
			wg.Wait()
		} else {

			dl.downloadChunk(info, file, 0, info.contentLength-1, nil, nil)
		}
		info.timeEnd = time.Now().Unix()
	}
}

func (dl *download) downloadChunk(info *downloadInfo, file *os.File, start int64, end int64, wg *sync.WaitGroup, fileLock *sync.Mutex)  {

	if wg != nil {
		defer wg.Done()
	}

	httpRequest, err := http.NewRequest(http.MethodGet, info.url, nil)
	if err != nil {
		info.err = err
		return
	}
	httpRequest.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", start, end))

	jar, err := cookiejar.New(nil)
	if err != nil {
		info.err = err
		return
	}

	timeStart := time.Now().UnixNano()
	httpClient := &http.Client{Jar:jar,}
	httpResponse, err := httpClient.Do(httpRequest)
	if err != nil {
		info.err = err
		return
	}
	defer httpResponse.Body.Close()

	dl.stats.setSpeed(end-start+1, time.Now().UnixNano() - timeStart)

	buf := dl.bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer dl.bufPool.Put(buf)

	writeIndex := int64(start)
	downloadSize := int64(0)
	for {
		n, err := buf.ReadFrom(httpResponse.Body)
		if n > 0 {
			if fileLock != nil {
				fileLock.Lock()
			}
			writeSize, err := file.WriteAt(buf.Bytes(), writeIndex)
			info.size += int64(writeSize)
			if fileLock != nil {
				fileLock.Unlock()
			}
			if err != nil {
				info.err = err
				return
			}
			writeIndex += int64(writeSize)

			dl.stats.setStats(info, time.Now().Unix() - info.timeBegin)
		}
		if err != nil {
			info.err = err
			return
		}
		downloadSize += n
		if downloadSize == end-start+1 {
			return
		}
	}
}