package wget

import (
	"fmt"
	"strings"
)

type downloadStats struct {

	templateStats string
	speed float32
}

func newDownloadStats() *downloadStats {

	return &downloadStats{
		//percent [===>      ] size speed time
		// 5% [===>   ] 1,235 19.8k/s 01m 21s
		templateStats:"\r%s\t[%s]\t%sBytes\t%0.2fKB/s\t%s",
		speed:0,
	}
}

func (s *downloadStats) setSpeed(nbytes int64, t int64)  {

	s.speed = float32(nbytes * 1000000000) / float32(t)
}

func (s *downloadStats) setStats(info *downloadInfo, t int64)  {

	percent := info.size * 100 / info.contentLength
	strPercent := fmt.Sprintf("%d", percent)
	strPercent += "%"

	gLog.logInfo(s.templateStats, strPercent, s.progress(info.size, info.contentLength), toStringI64(info.size), s.speed, formatTime(int(t), timeWithHMS))
}

func (s *downloadStats) progress(percent, total int64) string {

	equals := int(percent * 50 / total)
	spaces := 50 - equals

	prog := strings.Repeat("=", equals) + ">" + strings.Repeat(" ", spaces)
	return  prog
}