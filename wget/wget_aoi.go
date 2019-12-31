package wget

type aoi interface {

	getInterest(data string) (result []string)
	handleInterest()
	verbose(vb bool)
}