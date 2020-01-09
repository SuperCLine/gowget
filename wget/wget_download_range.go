package wget

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type downloadRange struct {

	mStart int64
	mEnd int64
}

func newDownloadRange(start, end int64) *downloadRange {

	return &downloadRange{
		mStart:start,
		mEnd:end,
	}
}

func (dr *downloadRange) HandleDownLoad(info *downloadFile, file *os.File, fileLocker *sync.Mutex, pool *sync.Pool) error {

	req, err := http.NewRequest(http.MethodGet, info.mUrl, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", dr.mStart, dr.mEnd))

	jar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}

	cli := &http.Client{
		Jar:jar,
		Timeout:time.Second * 180,
	}
	resp, err := cli.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return errors.New("Bad download range request.")
	}

	buf := pool.Get().(*bytes.Buffer)
	buf.Reset()
	defer pool.Put(buf)

	writeIndex := dr.mStart
	downSize := int64(0)
	for {
		n, err := buf.ReadFrom(resp.Body)
		if err != nil {
			return err
		}

		if n > 0 {

			if fileLocker != nil {
				fileLocker.Lock()
			}
			writeSize, err := file.WriteAt(buf.Bytes(), writeIndex)
			if fileLocker != nil {
				fileLocker.Unlock()
			}

			if err != nil {
				return err
			}

			writeIndex += int64(writeSize)
		}

		downSize += n
		atomic.AddInt64(&info.mSize, n)

		totalTime := time.Now().UnixNano() - info.mTimeBegin
		info.stats.SetDownloadStats(info.mFileName, info.mSize, info.mContentLength, totalTime)

		if downSize == dr.mEnd-dr.mStart+1 {
			return nil
		}
	}
}