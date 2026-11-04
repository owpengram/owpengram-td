package peers

import "github.com/owpengram/owpengram-td/tg"

type MsgAndEntities struct {
	Msg      string
	Entities []tg.MessageEntityClass
}
