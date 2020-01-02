package wget

import (
	"regexp"
	"strings"
)

type aoiHref struct {

	regHref *regexp.Regexp
	dl *download
	gg *gowget
}

func newAoiHref(gg *gowget) *aoiHref {

	aoi := &aoiHref{
		gg:gg,
		regHref:regexp.MustCompile(`<a href="(.*?)">`),
	}
	aoi.dl = newDownload(gg, aoi)

	return aoi
}

func (ah *aoiHref) GetInterest(data string) (result []string) {

	hrefAll := ah.regHref.FindAllStringSubmatch(data, -1)
	num := len(hrefAll)
	for i:=0; i<num; i++ {

		hrefUrl := hrefAll[i][1]
		if hrefUrl == ".." {
			continue
		}

		if isUrl(hrefUrl) {

			if ah.gg.flagForeignUrl {

				pos := strings.Index(hrefUrl, `"`)
				if pos != -1 {

					result = append(result, hrefUrl[:pos])
				} else {
					result = append(result, hrefUrl)
				}
			}
		} else {

			url := ah.getUrl(hrefUrl)
			pos := strings.Index(url, ah.gg.url)
			if pos == 0 {

				result = append(result, url)
			} else {

				if ah.gg.flagParentUrl {
					result = append(result, url)
				}
			}
		}
	}

	return
}

func (ah *aoiHref) HandleInterest()  {

	ah.dl.ParseDownLoad()
	ah.dl.HandleDownLoad()
}

func (ah *aoiHref) Verbose(vb bool)  {

	ah.dl.Verbose(vb)
}

func (ah *aoiHref) getUrl(hrefUrl string) string {

	url := ah.gg.baseUrl + "/" + hrefUrl
	url = formatUrl(url)

	return url
}