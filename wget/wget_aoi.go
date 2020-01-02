package wget

type aoi interface {

	GetInterest(data string) (result []string)
	HandleInterest()
	Verbose(vb bool)
}