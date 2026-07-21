package error

import (
	"fmt"
	"os"
)

type Reporter struct {
	callback func(error)
}

func (r *Reporter) Report(err error) {
	fmt.Fprintln(os.Stderr, err)
	r.callback(err)
}

func NewReporter(callback func(error)) *Reporter {
	return &Reporter{callback: callback}
}
