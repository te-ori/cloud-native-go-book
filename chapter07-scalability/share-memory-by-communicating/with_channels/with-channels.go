package with_channels

type Resource string

func Poller(in, out chan *Resource) {
	for r := range in {
		// Poll the url
		// ...

		// Send the processed Resource to out
		out <- r
	}
}
