package sessions

import "io"

func NewRandomSource(bytes int, randomSourceReader io.Reader) SessionIdGenerator {
	return &randomSource{
		reader: randomSourceReader,
		bytes:  bytes,
	}
}

// randomSource implements the SessionIdGenerator and creates random session IDs.
type randomSource struct {
	reader io.Reader
	bytes  int
}

func (r *randomSource) Generate() ([]byte, error) {
	b := make([]byte, r.bytes)
	_, err := r.reader.Read(b)
	return b, err
}
