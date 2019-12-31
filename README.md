# gowget
mini wget with golang, u can customize your aoi to grab what u want. aoi means anything of insterst, or area of insterst in game.

# [官网](https://supercline.com/game/tool-sdk/wget-with-go.html)
	https://supercline.com/game/tool-sdk/wget-with-go.html

# get code
	go get github.com/SuperCLine/gowget
	
# usage
    gowget version: gowget/1.0.0
    Usage: gowget [-VhrH] [-v -p -nv -np] [-o=path] url
    
    Options:
      -H	go to foreign hosts when recursive.
      -V	display the version of gowget.
      -h	print this help.
      -np	don't ascend to the parent directory.
      -nv	turn off verboseness, without being quiet.
      -o string
          	save directory.
      -p	include parent when recursive. (default true)
      -r	specify recursive download.
      -v	be verbose(this is the default). (default true)