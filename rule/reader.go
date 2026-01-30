package rule

import (
	"io"
)

// reusableReader stores the data read from the original reader so that it can
// be served again after EOF is reached. It resets the offset whenever EOF is
// returned allowing the body to be read multiple times.
type reusableReader struct {
	data []byte
	off  int
}

// ReusableReader reads the entire contents of r and returns a reader that can
// be consumed multiple times. Each time EOF is reached, the internal pointer is
// reset so subsequent reads start from the beginning again.
func ReusableReader(r io.Reader) io.Reader {
	data, _ := io.ReadAll(r)
	return &reusableReader{data: data}
}

func (r *reusableReader) Read(p []byte) (int, error) {
	if r.off >= len(r.data) {
		// Reset for next consumer and signal EOF.
		r.off = 0
		return 0, io.EOF
	}

	n := copy(p, r.data[r.off:])
	r.off += n
	if r.off >= len(r.data) {
		// Next read will start from the beginning.
		return n, io.EOF
	}

	return n, nil
}
