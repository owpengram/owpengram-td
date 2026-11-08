// Command layerimport imports a Telegram API schema Layer from an official
// Telegram Desktop revision into _schema/layers.
//
// It downloads Telegram/SourceFiles/mtproto/scheme/api.tl at the given
// tdesktop tag, branch or commit, checks that the file declares the expected
// "// LAYER N", stores it verbatim as layer-N.tl and records the resolved
// commit, git blob id and SHA-256 in manifest.json. With -canonical the Layer
// also becomes the manifest's canonical Layer (the one the generated tg
// package models).
//
// Typical use, after a new tdesktop release:
//
//	go run ./cmd/layerimport -ref v6.3.4 -canonical
//	go generate .
//
// policy.json is not touched: the reviewed decisions for a new Layer (dropped
// fields, rejected constructors, converters) still need a human pass, and
// gotdgen fails closed until they exist.
package main

import (
	"bytes"
	"crypto/sha1" // #nosec G505 -- git blob ids are SHA-1 by definition.
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/owpengram/owpengram-td/gen/semantic"
)

const (
	tdesktopRepo  = "telegramdesktop/tdesktop"
	tdesktopGit   = "https://github.com/telegramdesktop/tdesktop.git"
	apiSourcePath = "Telegram/SourceFiles/mtproto/scheme/api.tl"
)

var layerLine = regexp.MustCompile(`(?m)^//\s*LAYER\s+(\d+)\s*$`)

type layerEntry struct {
	Layer  int    `json:"layer"`
	File   string `json:"file"`
	Commit string `json:"commit"`
	Blob   string `json:"blob"`
	SHA256 string `json:"sha256"`
}

func main() {
	var (
		ref       = flag.String("ref", "", "tdesktop tag, branch or commit to import from (required)")
		layer     = flag.Int("layer", 0, "expected Layer number (default: whatever the schema declares)")
		dir       = flag.String("dir", filepath.Join("_schema", "layers"), "layers directory holding manifest.json")
		canonical = flag.Bool("canonical", false, "also make this Layer the manifest's canonical Layer")
		force     = flag.Bool("force", false, "replace an existing entry for the same Layer")
	)
	flag.Parse()
	if *ref == "" {
		fmt.Fprintln(os.Stderr, "layerimport: -ref is required (a tdesktop tag, branch or commit)")
		os.Exit(2)
	}
	if err := run(*ref, *layer, *dir, *canonical, *force); err != nil {
		fmt.Fprintln(os.Stderr, "layerimport:", err)
		os.Exit(1)
	}
}

func run(ref string, wantLayer int, dir string, canonical, force bool) error {
	client := &http.Client{Timeout: 2 * time.Minute}

	commit, err := resolveCommit(client, ref)
	if err != nil {
		return fmt.Errorf("resolve %q: %w", ref, err)
	}
	data, err := fetch(client, fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s", tdesktopRepo, commit, apiSourcePath))
	if err != nil {
		return fmt.Errorf("download api.tl at %s: %w", commit, err)
	}

	m := layerLine.FindSubmatch(data)
	if m == nil {
		return fmt.Errorf("api.tl at %s has no \"// LAYER N\" line", commit)
	}
	gotLayer, _ := strconv.Atoi(string(m[1]))
	if wantLayer != 0 && gotLayer != wantLayer {
		return fmt.Errorf("api.tl at %s declares Layer %d, want %d", commit, gotLayer, wantLayer)
	}

	manifestPath := filepath.Join(dir, "manifest.json")
	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		return err
	}
	// Decode into an ordered generic document so keys we do not know about are
	// written back untouched.
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(raw, &doc); err != nil {
		return fmt.Errorf("parse %s: %w", manifestPath, err)
	}
	if repo := strings.Trim(string(doc["repository"]), `"`); repo != "" && repo != tdesktopGit {
		return fmt.Errorf("manifest repository is %s, not %s", repo, tdesktopGit)
	}
	var layers []layerEntry
	if err := json.Unmarshal(doc["layers"], &layers); err != nil {
		return fmt.Errorf("parse manifest layers: %w", err)
	}

	entry := layerEntry{
		Layer:  gotLayer,
		File:   fmt.Sprintf("layer-%d.tl", gotLayer),
		Commit: commit,
		Blob:   gitBlobID(data),
		SHA256: sha256Hex(data),
	}
	replaced := false
	for i, l := range layers {
		if l.Layer != gotLayer {
			continue
		}
		if l.SHA256 == entry.SHA256 && l.Commit == entry.Commit {
			fmt.Printf("Layer %d is already imported from %s.\n", gotLayer, short(commit))
			if canonical {
				// Still make sure the merged canonical schema follows it.
				return rebuildCanonicalSchema(manifestPath, dir, data)
			}
			return nil
		}
		if !force {
			return fmt.Errorf("Layer %d is already imported from %s; pass -force to replace it", gotLayer, short(l.Commit))
		}
		layers[i] = entry
		replaced = true
	}
	if !replaced {
		layers = append(layers, entry)
	}
	sort.Slice(layers, func(i, j int) bool { return layers[i].Layer < layers[j].Layer })

	if err := os.WriteFile(filepath.Join(dir, entry.File), data, 0o644); err != nil {
		return err
	}
	if doc["layers"], err = json.Marshal(layers); err != nil {
		return err
	}
	if canonical {
		doc["canonical_layer"] = json.RawMessage(strconv.Itoa(gotLayer))
	}
	if err := writeManifest(manifestPath, raw, doc); err != nil {
		return err
	}
	if canonical {
		// The canonical Layer also owns _schema/telegram.tl (the merged schema
		// next to layers/). Rebuild it like gotdgen -schema-import does, or it
		// stays on the previous Layer while the generated code moves on.
		if err := rebuildCanonicalSchema(manifestPath, dir, data); err != nil {
			return err
		}
	}

	fmt.Printf("Imported Layer %d from tdesktop %s (%s)\n  file   %s\n  blob   %s\n  sha256 %s\n",
		gotLayer, ref, short(commit), filepath.Join(dir, entry.File), entry.Blob, entry.SHA256)
	if canonical {
		fmt.Printf("Canonical Layer is now %d.\n", gotLayer)
	}
	fmt.Println("Next: review _schema/layers/policy.json for this Layer, then run `go generate .`.")
	return nil
}

// rebuildCanonicalSchema writes ../telegram.tl from the updated manifest, the
// canonical Layer source and the manifest overlays.
func rebuildCanonicalSchema(manifestPath, dir string, canonicalSource []byte) error {
	manifest, err := semantic.ReadManifest(manifestPath)
	if err != nil {
		return err
	}
	overlays := make(map[string][]byte, len(manifest.Overlays))
	for _, overlay := range manifest.Overlays {
		data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(overlay.File)))
		if err != nil {
			return fmt.Errorf("read canonical overlay %q: %w", overlay.File, err)
		}
		overlays[overlay.File] = data
	}
	rendered, err := semantic.RenderCanonicalSchema(manifest, canonicalSource, overlays)
	if err != nil {
		return err
	}
	path := filepath.Clean(filepath.Join(dir, "..", "telegram.tl"))
	if err := os.WriteFile(path, rendered, 0o644); err != nil {
		return err
	}
	fmt.Printf("  canonical schema %s\n", path)
	return nil
}

// resolveCommit turns a tag, branch or (abbreviated) commit into a full commit id.
func resolveCommit(client *http.Client, ref string) (string, error) {
	body, err := fetch(client, fmt.Sprintf("https://api.github.com/repos/%s/commits/%s", tdesktopRepo, ref))
	if err != nil {
		return "", err
	}
	var out struct {
		SHA string `json:"sha"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", err
	}
	if len(out.SHA) != 40 {
		return "", fmt.Errorf("GitHub returned no commit id")
	}
	return out.SHA, nil
}

func fetch(client *http.Client, url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "owpengram-td-layerimport")
	if token := os.Getenv("GITHUB_TOKEN"); token != "" && strings.HasPrefix(url, "https://api.github.com/") {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s", url, resp.Status)
	}
	return body, nil
}

// writeManifest rewrites manifest.json keeping the original key order and
// two-space indentation, so a Layer import is a small reviewable diff.
func writeManifest(path string, original []byte, doc map[string]json.RawMessage) error {
	keys, err := topLevelKeys(original)
	if err != nil {
		return err
	}
	seen := map[string]bool{}
	var buf bytes.Buffer
	buf.WriteString("{\n")
	first := true
	emit := func(k string) error {
		var v bytes.Buffer
		if err := json.Indent(&v, doc[k], "  ", "  "); err != nil {
			return err
		}
		if !first {
			buf.WriteString(",\n")
		}
		first = false
		name, _ := json.Marshal(k)
		fmt.Fprintf(&buf, "  %s: %s", name, v.Bytes())
		return nil
	}
	for _, k := range keys {
		seen[k] = true
		if err := emit(k); err != nil {
			return err
		}
	}
	var extra []string
	for k := range doc {
		if !seen[k] {
			extra = append(extra, k)
		}
	}
	sort.Strings(extra)
	for _, k := range extra {
		if err := emit(k); err != nil {
			return err
		}
	}
	buf.WriteString("\n}\n")
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

func topLevelKeys(data []byte) ([]string, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	if _, err := dec.Token(); err != nil {
		return nil, err
	}
	var keys []string
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		keys = append(keys, tok.(string))
		var skip json.RawMessage
		if err := dec.Decode(&skip); err != nil {
			return nil, err
		}
	}
	return keys, nil
}

func gitBlobID(data []byte) string {
	h := sha1.New() // #nosec G401 -- git object id, not a security primitive.
	fmt.Fprintf(h, "blob %d\x00", len(data))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func short(commit string) string {
	if len(commit) > 10 {
		return commit[:10]
	}
	return commit
}
