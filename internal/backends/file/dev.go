package file

import (
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"sort"
	"time"

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
	tailer    *tail.Tail
	linesCh   chan string
	pathCh    chan string
	watcher   *watcher.Watcher
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
	h.dir, _ = filepath.Abs(dir)
	h.dir += "/"

	// set files mask (default is *.log)
	mask := sconf.StrDef(h.confScope, "mask", `.*\.log`)
	if h.mask, err = regexp.Compile(mask); err != nil {
		return err
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

func (h *handler) read() (string, error) {

	select {
	case line, ok := <-h.linesCh:
		if ok {
			return line, nil
		}
	}

	return "", nil
}

func (h *handler) write(s string) error {
	return fmt.Errorf("not implemented")
}

func (h *handler) stop() {
	// stop dir watcher
	h.watcher.Close()

	// stop tailer
	h.tailer.Stop()
	h.tailer.Cleanup()

	close(h.linesCh)
	close(h.pathCh)
}

func (h *handler) getRecentFile() string {
	if len(h.files) > 0 {
		sort.Strings(h.files)
		return h.files[len(h.files)-1]
	}

	return ""
}

func (h *handler) watchDir() error {

	// monitor the direcotory for newer file to appear

	h.watcher = watcher.New()
	h.watcher.FilterOps(watcher.Create)
	h.watcher.AddFilterHook(watcher.RegexFilterHook(h.mask, false))

	if err := h.watcher.Add(h.dir); err != nil {
		return err
	}

	for path, _ := range h.watcher.WatchedFiles() {
		h.files = append(h.files, path)
	}

	if p := h.getRecentFile(); p != "" {
		h.pathCh <- p
	}

	// start dir watcher
	go func() {
		for {
			select {
			case event := <-h.watcher.Event:
				h.files = append(h.files, event.Path)
				if rf := h.getRecentFile(); rf != "" {
					h.pathCh <- h.getRecentFile()
				}
			case err := <-h.watcher.Error:
				slog.Err("%v\n", err)
			case <-h.watcher.Closed:
				return
			}
		}
	}()

	go h.watcher.Start(time.Second * 1)

	return nil
}

func (h *handler) tailFile() {
	var err error

	cfg := tail.Config{
		ReOpen: true,
		Follow: true,
		Poll:   true,
		Location: &tail.SeekInfo{
			Offset: 0,
			Whence: io.SeekEnd},
	}

	// we must have smth to start with in case of target file absence
	h.tailer, _ = tail.TailFile("/dev/null", cfg)

	pathChanged := false

	for {

		select {

		case path, ok := <-h.pathCh:

			if !ok {
				break
			}

			slog.Debug(5, "tailer: watching '%s'\n", path)

			h.tailer.Stop()
			h.tailer.Cleanup()

			h.tailer, err = tail.TailFile(path, cfg)

			if !pathChanged {
				cfg.Location.Whence = io.SeekStart
				pathChanged = true
			}

			if err != nil {
				slog.Err("tailer: %v\n", err)
			}

		case line, ok := <-h.tailer.Lines:

			if !ok {
				continue
			}

			h.linesCh <- line.Text

		} //select

	} //for

}
