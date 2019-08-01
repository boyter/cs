package processor

// FileJob is a struct used to hold all of the results of processing internally before sent to the formatter
type FileJob struct {
	Filename  string
	Extension string
	Location  string
	Content   []byte
	Bytes     int64
	Hash      []byte
	Binary    bool
	Score     float64
	Locations map[string][]int
}
