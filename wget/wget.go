package wget

import (
	"flag"
	"os"
	"strings"
)

const (
	Ver = "gowget/1.0.0"
)

type gowget struct {

	flagVersion bool
	flagHelp bool
	flagVerbose bool
	flagRecursive bool
	flagForeignUrl bool
	flagParentUrl bool
	flagOutpath string

	flagInclude []string
	flagExclude []string

	url string
	baseUrl string

	instHref aoi
}

func New() *gowget {

	return &gowget{

	}
}

func (gg *gowget) Run(args []string) {

	if gg.parseArgs(args) {

		gg.instHref.HandleInterest()
		gg.instHref.Verbose(gg.flagVerbose)
	}
}

func (gg *gowget) usage()  {

	gLog.logInfo(`gowget version: gowget/1.0.0
Usage: gowget [-VhrH] [-v -p -nv -np] [-o=path] url

Options:
`)
	flag.PrintDefaults()
}

func (gg *gowget) parseArgs(args []string) bool {

	flag.BoolVar(&gg.flagVersion, "V", false, "display the version of gowget.")
	flag.BoolVar(&gg.flagHelp, "h", false, "print this help.")
	flag.BoolVar(&gg.flagRecursive, "r", false, "specify recursive download.")
	flag.BoolVar(&gg.flagForeignUrl, "H", false, "go to foreign hosts when recursive.")

	flag.BoolVar(&gg.flagVerbose, "nv", false, "turn off verboseness, without being quiet.")
	flag.BoolVar(&gg.flagParentUrl, "np", false, "don't ascend to the parent directory.")
	flag.BoolVar(&gg.flagVerbose, "v", true, "be verbose(this is the default).")
	flag.BoolVar(&gg.flagParentUrl, "p", true, "include parent when recursive.")

	flag.StringVar(&gg.flagOutpath, "o", "", "save directory.")

	inc := flag.String("I", "", "list of allowed directories.")
	exc := flag.String("X", "", "list of excluded directories.")

	flag.Usage = gg.usage

	flag.Parse()

	if *inc != "" {
		gg.flagInclude = strings.Split(*inc, ",")
		for i:=0; i<len(gg.flagInclude); i++ {
			gg.flagInclude[i] = strings.Trim(gg.flagInclude[i], " ")
		}
	}

	if *exc != "" {
		gg.flagExclude = strings.Split(*exc, ",")
		for i:=0; i<len(gg.flagExclude); i++ {
			gg.flagExclude[i] = strings.Trim(gg.flagExclude[i], " ")
		}
	}

	numArgs := len(args)
	if gg.flagHelp || numArgs == 0 {

		flag.Usage()
		return false
	} else if gg.flagVersion {

		gLog.logInfo("%s", Ver)
		return false
	} else {

		gg.url = args[numArgs - 1]
		if !isUrl(gg.url) {
			gLog.logErr("url[%s] is not support.", gg.url)
			return false
		}

		pos := indexUrl(gg.url, 3)
		if pos == -1 {
			gg.baseUrl = gg.url
		} else {
			gg.baseUrl = gg.url[:pos]
		}
		gg.baseUrl += "/"

		if gg.flagOutpath == "" {
			gg.flagOutpath, _ = os.Getwd()
		}

		if err := createDirectory(gg.flagOutpath); err != nil {
			gLog.logErr("failed to create out path.")
			return false
		}

		gg.instHref = newAoiHref(gg)

		return true
	}
}

func (gg *gowget) testUrl(url string) bool {

	if len(gg.flagInclude) > 0 {

		for i:=0; i<len(gg.flagInclude); i++ {
			if strings.Contains(url, gg.flagInclude[i]) {
				return true
			}
		}

		return false
	}

	if len(gg.flagExclude) > 0 {

		for i:=0; i<len(gg.flagExclude); i++ {
			if strings.Contains(url, gg.flagExclude[i]) {
				return false
			}
		}

		return true
	}

	return true
}