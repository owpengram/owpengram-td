package tlprofile

import (
	"bytes"
	"testing"

	"github.com/owpengram/owpengram-td/bin"
	"github.com/owpengram/owpengram-td/tg"
)

func TestLayer233CanonicalProfile(t *testing.T) {
	if tg.Layer != 233 || ProfileCanonical != Profile233 {
		t.Fatalf("canonical Layer = tg:%d profile:%d, want 233", tg.Layer, ProfileCanonical)
	}
	if tg.ChannelsToggleForumRequestTypeID != 0xa2952b54 {
		t.Fatalf("channels.toggleForum wire ID = %#x, want 0xa2952b54", tg.ChannelsToggleForumRequestTypeID)
	}
}

func TestLayer233HistoricalToggleForumDefaultsSidebar(t *testing.T) {
	// channels.toggleForum#3ff75734 channel:InputChannel enabled:Bool tabs:Bool
	var wire bin.Buffer
	wire.PutID(0x3ff75734)
	wire.PutID(tg.InputChannelEmptyTypeID)
	wire.PutID(tg.BoolTrueTypeID)
	wire.PutID(tg.BoolTrueTypeID)

	for profile := Profile225; profile < ProfileCanonical; profile++ {
		in := &bin.Buffer{Buf: append([]byte(nil), wire.Buf...)}
		got, err := DecodeObject(profile, in, Limits{})
		if err != nil {
			t.Fatalf("%d decode historical toggleForum: %v", profile, err)
		}
		request, ok := got.(*tg.ChannelsToggleForumRequest)
		if !ok {
			t.Fatalf("%d decoded %T, want *tg.ChannelsToggleForumRequest", profile, got)
		}
		if !request.Enabled || !request.Tabs || request.Sidebar {
			t.Fatalf("%d decoded %+v, want enabled+tabs without sidebar", profile, request)
		}
	}
}

func TestLayer233SidebarFlagsDropForHistoricalProfiles(t *testing.T) {
	for _, pair := range []struct {
		name       string
		with, bare bin.Object
	}{
		{"user", &tg.User{ID: 1, Bot: true, BotForumSidebar: true}, &tg.User{ID: 1, Bot: true}},
		{"channel", &tg.Channel{ID: 1, Title: "c", Photo: &tg.ChatPhotoEmpty{}, ForumSidebar: true}, &tg.Channel{ID: 1, Title: "c", Photo: &tg.ChatPhotoEmpty{}}},
	} {
		for profile := Profile225; profile < ProfileCanonical; profile++ {
			with := layer232Encode(t, profile, pair.with)
			bare := layer232Encode(t, profile, pair.bare)
			if !bytes.Equal(with, bare) {
				t.Fatalf("%d %s keeps the Layer 233 sidebar flag", profile, pair.name)
			}
		}
		with := layer232Encode(t, ProfileCanonical, pair.with)
		bare := layer232Encode(t, ProfileCanonical, pair.bare)
		if bytes.Equal(with, bare) {
			t.Fatalf("canonical %s lost the sidebar flag", pair.name)
		}
	}
}
