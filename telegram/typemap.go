package telegram

import (
	"sync"

	"github.com/owpengram/owpengram-td/mt"
	"github.com/owpengram/owpengram-td/proto"
	"github.com/owpengram/owpengram-td/tg"
	"github.com/owpengram/owpengram-td/tmap"
)

// Port is default port used by telegram.
const Port = 443

var (
	typesMap  *tmap.Map
	typesOnce sync.Once
)

func getTypesMapping() *tmap.Map {
	typesOnce.Do(func() {
		typesMap = tmap.New(
			tg.TypesMap(),
			mt.TypesMap(),
			proto.TypesMap(),
		)
	})
	return typesMap
}
