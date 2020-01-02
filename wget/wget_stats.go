package wget

import (
	"fmt"
	"strings"
)

type downloadStats struct {

	templateStats string
}

func newDownloadStats() *downloadStats {

	return &downloadStats{
		//percent [===>      ] size speed time
		// 5% [===>   ] 1,235 19.8k/s 01m 21s
		templateStats:"\r%s\t[%s]\t%sBytes\t%0.2fKB/s\t%s\t[%s]",
	}
}

func (s *downloadStats) SetStats(fileName string, size, totalSize int64, t int64)  {

	percent := size * 100 / totalSize
	strPercent := fmt.Sprintf("%d", percent)
	strPercent += "%"

	speed := float32(size * 1000000) / float32(t)

	gLog.logInfo(s.templateStats, strPercent, s.progress(size, totalSize), toStringI64(size), speed, formatTime(int(t/1000000000), timeWithHMS), fileName)
}

func (s *downloadStats) progress(size, totalSize int64) string {

	equals := int(size * 50 / totalSize)
	spaces := 50 - equals
	prog := strings.Repeat("=", equals) + ">" + strings.Repeat(" ", spaces)

	return prog
}