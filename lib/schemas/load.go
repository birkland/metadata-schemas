//go:generate go run github.com/gobuffalo/packr/v2/packr2

package schemas

import (
	"io"

	"github.com/gobuffalo/packr/v2"
)

func Load(id string) (io.ReadCloser, error) {
	box := packr.New("myBox", "../../schemas")
	return box.Open(id)
}
