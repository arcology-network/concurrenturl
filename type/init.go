package urltype

import (
	"encoding/gob"
)

func init() {
	gob.Register(&Univalue{})
}
