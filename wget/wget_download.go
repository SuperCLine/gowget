package wget

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	numThread = 100
)

type download struct {

	aoi aoi
	gg *gowget
	client *http.Client
	dlList []*downloadFile
	stats *downloadStats
}

func newDownload(gg *gowget, aoi aoi) *download {

	return &download{
		gg:gg,
		aoi:aoi,
		stats:newDownloadStats(),
		client:&http.Client{Timeout:time.Second * 180},
		dlList:make([]*downloadFile, 0),
	}
}

func (dl *download) ParseDownLoad()  {

	gLog.logInfo("\rBegin parsing, pls waiting ...")
	dl.parseHead(dl.gg.url, true)
	gLog.logInfo("\rEnd parsing.")
}

func (dl *download) HandleDownLoad() {

	numFile := len(dl.dlList)
	num := clamp(numFile, 0, numThread)

	wq := NewWorkQueue()
	wq.Init(num, numFile)

	for i:=0; i<numFile; i++ {

		downFile := dl.dlList[i]
		wq.AddTask(NewDefaultWork(func() {

			downFile.HandleDownLoad()
		}))
	}

	wq.Destroy()
}

func (dl *download) Verbose(vb bool)  {

	if dl.gg.flagVerbose {

		succeedNum := 0
		failedNum := 0
		for i:=0; i<len(dl.dlList); i++ {

			downFile := dl.dlList[i]

			strErr := "nil"
			if downFile.mErr != nil {
				strErr = downFile.mErr.Error()
				failedNum++
			} else {
				succeedNum++
			}

			gLog.logInfo("\ndownload:%s, error:%s, time:%s",
				downFile.mUrl,
				strErr,
				formatTime(int((downFile.mTimeEnd-downFile.mTimeBegin)/1000000000), timeWithHMS))

			// TO DO: save range to json file for resume from break point
		}

		gLog.logInfo("\nSucceed:%d, Failed:%d", succeedNum, failedNum)
	}
}

func (dl *download) isRedirectGet(statusCode int) bool {

	switch statusCode {
	case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther, http.StatusTemporaryRedirect:
		return true
	}

	return false
}

func (dl *download) httpRequest(method, url string) (*http.Response, error) {

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		gLog.logErr("failed to new request.[err:%s, url:%s]", err.Error(), url)
		return nil, err
	}

	resp, err := dl.client.Do(req)
	if err != nil {
		gLog.logErr("failed to get response.[err:%s, url:%s]", err.Error(), url)
		return nil, err
	}

	return resp, nil
}

func (dl *download) parseHead(url string, recursive bool)  {

	headResp, err := dl.httpRequest(http.MethodHead, url)
	if err != nil {
		return
	}
	defer headResp.Body.Close()

	cType := headResp.Header.Get("Content-Type")
	if strings.Contains(cType, "html") {

		if recursive {

			htmlResp, err := dl.httpRequest(http.MethodGet, url)
			if err != nil {
				return
			}
			defer htmlResp.Body.Close()

			body, err := ioutil.ReadAll(htmlResp.Body)
			if err != nil {
				gLog.logErr("failed to read html.[err:%s, url:%s]", err.Error(), url)
			} else {
				dl.writeHtml(url, string(body))
				dl.parseHtml(string(body), dl.gg.flagRecursive)
			}
		}
	} else {

		if headResp.StatusCode == 200 {

			dl.addDownloadFile(headResp, url, cType)
		} else {

			if dl.isRedirectGet(headResp.Request.Response.StatusCode) {

				fileResp, err := dl.httpRequest(http.MethodGet, url)
				if err != nil {
					return
				}
				defer fileResp.Body.Close()

				cType = fileResp.Header.Get("Content-Type")

				dl.addDownloadFile(fileResp, url, cType)
			} else {

				gLog.logErr("\nfailed to get file data.[code:%d, url:%s]", headResp.StatusCode, url)
			}
		}

	}

	dl.stats.SetParseStats()
}

func (dl *download) addDownloadFile(resp *http.Response, url, cType string)  {

	rangeBytes := resp.Header.Get("Accept-Ranges")
	acceptRange := false
	if strings.Compare(rangeBytes, "bytes") == 0 {
		acceptRange = true
	}

	dlFile, err := newDownloadFile(dl.gg.flagOutpath, url, acceptRange, resp.ContentLength, cType, dl.stats)
	if err != nil {

		gLog.logErr("fail to new download file. err:%s", err.Error())
	} else {
		dl.dlList = append(dl.dlList, dlFile)
	}
}

func (dl *download) writeHtml(url, body string)  {

	dir, filePath := getWritePath(dl.gg.flagOutpath, url, true)
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

	result := dl.aoi.GetInterest(body)
	num:=len(result)
	for i:=0; i<num; i++ {

		dl.parseHead(result[i], recursive)
	}
}

