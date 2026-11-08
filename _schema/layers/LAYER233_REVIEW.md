# Layer 233 schema review

Source: official Telegram Desktop `v7.4.1` (stable, `AppBetaVersion = false`),
commit `157f5cfc85`. `manifest.json` locks the Git blob
`4e29f2deb45a083196a8c7ac09cdedf2600c1a10` and normalized SHA-256
`c2bbefcaec3285ba46c8af8f429c7012fe5aea253fd16d4be089f91c7c3fa71e`.
Imported with `go run ./cmd/layerimport -ref v7.4.1 -canonical`. No client
patch is part of the schema import.

Relative to Layer 232 the schema adds 31 constructors and changes one
constructor ID. Canonical moves to 233; exact profiles 225–232 remain.

## Constructor changes

| Constructor | Layer 232 | Layer 233 |
|---|---|---|
| `channels.toggleForum` | `#3ff75734` | `#a2952b54` (adds required `sidebar:Bool`) |
| `user` | `#b1b8cc83` | `#b1b8cc83` (adds `bot_forum_sidebar:flags2.22?true`, same ID) |
| `channel` | `#7ad15240` | `#7ad15240` (adds `forum_sidebar:flags2.23?true`, same ID) |

New constructors: message widgets (`inputMediaWidget`, `inputMediaWidgetId`,
`messageMediaWidget`, `inputBotInlineMessageWidget`,
`inputBotInlineMessageWidgetId`, `botInlineMessageWidget`, `userWallet`,
`widgetQueryTypeData`, `widgetQueryTypeSignData`,
`widgetQueryTypeSentTransaction`, `messages.botWidgetQueryAnswer`,
`messages.widgetPublicDataNotModified`, `messages.widgetPublicData` and the
methods `messages.getBotWidgetQueryAnswer`, `messages.setBotWidgetQueryAnswer`,
`messages.editMessageWidgetData`, `messages.editInlineBotMessageWidgetData`,
`messages.editWidgetData`, `messages.getWidgetPublicData`), audio file
transcription (`messages.transcribeAudioFile`, `messages.getTranscriptionLangs`)
and the updates `updateEditForumTopic`, `updateBotWidgetQuery`,
`updateInlineBotWidgetQuery`, `updateBotWidgetCopyQuery`,
`updateMessageWidgetData` (carries `pts`/`pts_count`) and
`updateTranscribedAudioFile`.

The MTProto service schema is unchanged.

## Compatibility decisions

The policy audit found 56 new obligations: seven decisions for each historical
profile 225–232.

| Scenario | Decision |
|---|---|
| Historical `channels.toggleForum#3ff75734` (225–232) | `default`: the missing required `sidebar` becomes `false`. A historical client has no forum sidebar, so turning a forum on or off from it leaves the sidebar off. |
| `updateEditForumTopic` encoded for 225–232 | `drop`. No `pts`; historical clients still learn about topic edits from the service message. |
| `updateTranscribedAudioFile` encoded for 225–232 | `drop`. It only answers `messages.transcribeAudioFile`, which a historical client cannot call. |
| `updateBotWidgetQuery`, `updateInlineBotWidgetQuery`, `updateBotWidgetCopyQuery` encoded for 225–232 | `reject`. They can carry wallet sign requests and sent transactions; as with the Layer 230 wallet updates, financial and session events are never dropped silently. A server must not send them to historical bots. |
| `updateMessageWidgetData` encoded for 225–232 | `reject`. It carries `pts`/`pts_count`: dropping it would leave a pts gap in a historical client. A server must not emit it towards historical profiles. |
| `user.bot_forum_sidebar`, `channel.forum_sidebar` encoded for 225–232 | Drop (generic optional-field default). Display flags only, the constructor ID is unchanged. |
| New constructors in historical profiles | Unavailable (generic default). |

The reviewed policy document now holds 305 entries (249 retained, 56 new, zero
stale or unresolved).

## 2026-11-08 verification

On Windows/amd64:

- Schema import verified against the locked Git blob and normalized SHA-256;
  `canonical_layer` moved to 233 and `layer-233.tl` was written with its
  provenance.
- `go generate .` completed with the reviewed policy; `tg.Layer = 233`.
- `cmd/layerimport` now also rebuilds the merged `_schema/telegram.tl` when
  it makes a Layer canonical (it used to leave it on the previous Layer).
- The policy is written by `TestRealLayerPolicy` (`UPDATE_LAYER_POLICY=1`);
  its Layer 233 rules live in `gen/layer_real_policy_test.go` and reproduce the
  reviewed document byte for byte.
- Locked generator budgets re-measured for the ninth profile: 743 projected
  families, 2888 rewrite variants, 1181 unavailable variants, 266 body
  obligations, 2172/1193/1102 execution plans (caps 2240/1230/1140), 1544
  typed codec wires (cap 1560) and 399 unprofiled wrapper invariants.
- New `tlprofile/layer233_test.go`: canonical 233 and the new
  `channels.toggleForum` ID, historical toggleForum decoding with
  `sidebar = false` in 225–232, and the sidebar flags dropped for 225–232
  but kept in the canonical profile.
- `go test ./...` (76 packages) and `go vet` pass.
