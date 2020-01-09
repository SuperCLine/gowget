package wget

import (
	"bytes"
	"net/url"
	"os"
	"runtime"
	"sync"
	"time"
)

const (
	gPackageSize = 64 * 1024
	gReceiveSize = 4 * 1024
	gSubSliceNum = 16
)

type downloadFile struct {

	mAcceptRange bool
	mContentLength int64
	mContentType string

	mUrl string
	mFileName string
	mFilePath string

	mTimeBegin int64
	mTimeEnd   int64

	mSize int64
	mErr error

	stats *downloadStats
	mFailRanges []*downloadRange
}

func newDownloadFile(outPath string, downloadUrl string, acceptRange bool, contentLength int64, contentType string, stats *downloadStats) (*downloadFile, error) {

	f := &downloadFile{
		mUrl:downloadUrl,
		mAcceptRange:acceptRange,
		mContentLength:contentLength,
		mContentType:contentType,
		mTimeBegin:0,
		mTimeEnd:0,
		mSize:0,
		mErr:nil,
		stats:stats,
		mFailRanges:make([]*downloadRange, 0),
	}

	normalUrl, err := url.PathUnescape(downloadUrl)
	if err != nil {
		return nil, err
	}

	_, f.mFilePath = getWritePath(outPath, normalUrl, false)
	f.mFileName = getFileName(f.mFilePath)

	createDirectory2(f.mFilePath)

	return f, nil
}

func (di *downloadFile) HandleDownLoad() {

	file, err := os.Create(di.mFilePath)
	if err != nil {
		di.mErr = err
		return
	}
	defer file.Close()

	fileLocker := &sync.Mutex{}
	bufPool := &sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, gReceiveSize))
		},
	}

	numPackage := int(di.mContentLength / gPackageSize)
	modPackage := di.mContentLength % gPackageSize
	if modPackage > 0 {
		numPackage += 1
	}

	di.mTimeBegin = time.Now().UnixNano()
	if di.mAcceptRange {

		failRanges := make([]*downloadRange, 0)
		wq := NewWorkQueue()
		wq.Init(runtime.NumCPU() * 2 + 2, numPackage)

		for i:=0; i<numPackage; i++ {

			start := int64(i) * gPackageSize
			end := start + gPackageSize
			if i == numPackage - 1 {
				end = di.mContentLength
			}

			dr := newDownloadRange(start, end-1)
			wq.AddTask(NewDefaultWork(func() {

				err = dr.HandleDownLoad(di, file, fileLocker, bufPool)
				if err != nil {
					failRanges = append(failRanges, dr)
				}
			}))
		}

		wq.WaitAllTask()

		lastFailRanges := make([]*downloadRange, 0)
		if len(failRanges) > 0 {

			for i:=0; i<len(failRanges); i++ {

				sliceSize := (failRanges[i].mEnd - failRanges[i].mStart) / gSubSliceNum
				sliceNum := gSubSliceNum
				if failRanges[i].mEnd - failRanges[i].mStart < gPackageSize - 1 {
					sliceSize = failRanges[i].mEnd - failRanges[i].mStart
					sliceNum =  1
				}
				for j:=0; j<sliceNum; j++ {

					start := failRanges[i].mStart + int64(j) * sliceSize
					end := start + sliceSize
					if j == sliceNum - 1 {
						end = failRanges[i].mEnd
					}

					dr := newDownloadRange(start, end)
					wq.AddTask(NewDefaultWork(func() {

						err = dr.HandleDownLoad(di, file, fileLocker, bufPool)
						if err != nil {
							lastFailRanges = append(lastFailRanges, dr)
						}
					}))
				}
			}
		}

		wq.WaitAllTask()

		for i:=0; i<len(lastFailRanges); i++ {

			err = lastFailRanges[i].HandleDownLoad(di, file, fileLocker, bufPool)
			if err != nil {
				di.mErr = err
				di.mFailRanges = lastFailRanges[i:]
				break
			}
		}

		wq.Destroy()

	} else {

		dr := newDownloadRange(0, di.mContentLength - 1)
		err := dr.HandleDownLoad(di, file, fileLocker, bufPool)
		if err != nil {
			di.mErr = err
			di.mFailRanges = append(di.mFailRanges, dr)
		}
	}
	di.mTimeEnd = time.Now().UnixNano()
}