package file

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	//pb "gamenode/pkg/gamenodepb"

	"github.com/maxb-odessa/sconf"
	"github.com/maxb-odessa/slog"
	"github.com/nxadm/tail"
	"github.com/radovskyb/watcher"
)

type handler struct {
	confScope string
	files     []string
	dir       string
	mask      *regexp.Regexp
	linesCh   chan string
	pathCh    chan string
}

func newDev(confScope string) *handler {
	return &handler{
		confScope: confScope,
	}
}

func (h *handler) run() error {
	var err error

	// set files dir (default is current)
	dir := sconf.StrDef(h.confScope, "dir", "./")
	h.dir, _ = filepath.Abs(os.ExpandEnv(dir))
	h.dir += "/"

	// set files mask (default is *.log)
	mask := sconf.StrDef(h.confScope, "mask", `.*\.log`)
	if h.mask, err = regexp.Compile(mask); err != nil {
		return err
	}

	mode := sconf.StrDef(h.confScope, "mode", "read")
	if mode != "read" {
		return fmt.Errorf("unsupported mode '%s'", mode)
	}

	h.linesCh = make(chan string, 32) // TODO adjust buf size here
	h.pathCh = make(chan string, 1)

	// start dir watcher
	if err := h.watchDir(); err != nil {
		return err
	}

	// start tailer
	go h.tailFile()

	return nil
}

func (h *handler) read() (interface{}, error) {

	select {
	case line, ok := <-h.linesCh:
		if ok {
			return line, nil
		}
	}

	return nil, nil
}

func (h *handler) write(i interface{}) error {
	return fmt.Errorf("not implemented")
}

/*
func (h *handler) stop() {
	// stop dir watcher
	h.watcher.Close()

	// stop tailer
	h.tailer.Stop()
	h.tailer.Cleanup()

	close(h.linesCh)
	close(h.pathCh)
}
*/

func (h *handler) getRecentFile() string {

	sortAsc := true

	if sconf.StrDef(h.confScope, "sort", "asc") != "asc" {
		sortAsc = false
	}

	if len(h.files) > 0 {
		sort.Strings(h.files)
		if sortAsc {
			return h.files[len(h.files)-1]
		} else {
			return h.files[0]
		}
	}

	return ""
}

func (h *handler) watchDir() error {

	// monitor the direcotory for newer file to appear

	w := watcher.New()
	w.FilterOps(watcher.Create)
	w.AddFilterHook(watcher.RegexFilterHook(h.mask, false))

	if err := w.Add(h.dir); err != nil {
		return err
	}

	for path, _ := range w.WatchedFiles() {
		h.files = append(h.files, path)
	}

	if p := h.getRecentFile(); p != "" {
		h.pathCh <- p
	}

	// start dir watcher
	go func() {
		for {
			select {
			case event := <-w.Event:
				h.files = append(h.files, event.Path)
				if rf := h.getRecentFile(); rf != "" {
					h.pathCh <- h.getRecentFile()
				}
			case err := <-w.Error:
				slog.Err("file watcher error: %s", err)
			case <-w.Closed:
				return
			}
		}
	}()

	go w.Start(time.Second * 1)

	return nil
}

func (h *handler) tailFile() {

	ctx, cancel := context.WithCancel(context.Background())

	for {

		select {
		case path, ok := <-h.pathCh:
			if !ok {
				break
			}
			cancel()
			ctx, cancel = context.WithCancel(context.Background())
			go h.realTailer(ctx, path)
		} //select

	} //for

}

func (h *handler) realTailer(ctx context.Context, path string) {

	cfg := tail.Config{
		ReOpen: true,
		Follow: true,
		Poll:   true,
		Location: &tail.SeekInfo{
			Offset: 0,
			Whence: io.SeekEnd,
		},
	}

	slog.Debug(5, "tailer: tailing '%s'", path)
	tailer, err := tail.TailFile(path, cfg)
	if err != nil {
		slog.Err("tailer failed on '%s': %s", path, err)
	}

	defer tailer.Stop()
	defer tailer.Cleanup()
	defer slog.Debug(5, "tailer: stopped tailing '%s'", path)

	for {
		select {
		case <-ctx.Done():
			return
		case line := <-tailer.Lines:
			h.linesCh <- line.Text
		}
	}
}
