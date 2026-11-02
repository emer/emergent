// Copyright (c) 2026, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package egui

import (
	"slices"
	"sync"
	"time"

	"cogentcore.org/core/core"
	"cogentcore.org/core/styles"
	"github.com/emer/emergent/v2/looper"
)

// Destroyer is a sim with resources to release when its GUI is destroyed,
// such as those of the GPU, which [Embed] calls once the sim has stopped
// running. The GPU state of a sim is typically global, as that of axon,
// so it must be released before the next sim is made.
type Destroyer interface {
	Destroy()
}

// releaseWait is the longest time that a sim whose GUI is destroyed is
// waited for to stop running before it is released anyway.
const releaseWait = 5 * time.Second

// releasing has the sims whose GUI is destroyed and that are being
// released once they stop running, which [Embed] waits for before it
// makes the next sim, so that no two sims use the global GPU state.
var releasing struct {
	sync.Mutex
	sims []chan struct{} // closed when released
}

// simFrame is the frame that [Embed] puts the GUI of a sim in, which stops
// and releases the sim when it is destroyed, as when the page of a content
// app that shows it is left.
type simFrame struct { //core:no-new
	core.Frame

	// gui is the GUI of the sim, set by [GUI.MakeBody].
	gui *GUI

	// sim is the sim, which is released if it is a [Destroyer].
	sim any

	// released is closed once the sim is released.
	released chan struct{}
}

func (sf *simFrame) Init() {
	sf.Frame.Init()
	sf.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Grow.Set(1, 1)
	})
}

func (sf *simFrame) Destroy() {
	if sf.released != nil {
		sf.release()
	}
	sf.Frame.Destroy()
}

// release stops the sim, and then releases it: now if it is not running,
// and otherwise once it stops, in another goroutine, so that the event
// loop, which its GUI updates need, is not blocked.
func (sf *simFrame) release() {
	released := sf.released
	sf.released = nil
	done := func() {
		if d, ok := sf.sim.(Destroyer); ok {
			d.Destroy()
		}
		close(released)
	}
	gui := sf.gui
	if gui == nil || !gui.IsRunning() {
		done()
		return
	}
	gui.stopRunning()
	releasing.Lock()
	releasing.sims = append(releasing.sims, released)
	releasing.Unlock()
	go func() {
		for start := time.Now(); gui.IsRunning() && time.Since(start) < releaseWait; {
			time.Sleep(10 * time.Millisecond)
		}
		done()
		releasing.Lock()
		releasing.sims = slices.DeleteFunc(releasing.sims, func(c chan struct{}) bool { return c == released })
		releasing.Unlock()
	}()
}

// pendingReleases returns the channels of the sims that are being released.
func pendingReleases() []chan struct{} {
	releasing.Lock()
	defer releasing.Unlock()
	return slices.Clone(releasing.sims)
}

// addLooper records the given loops, to stop them when the GUI is destroyed.
func (gui *GUI) addLooper(loops *looper.Stacks) {
	gui.runMu.Lock()
	defer gui.runMu.Unlock()
	if !slices.Contains(gui.loopers, loops) {
		gui.loopers = append(gui.loopers, loops)
	}
}

// stopRunning stops the running loops of the GUI at their innermost level,
// and sets the stop flag for other running processes.
func (gui *GUI) stopRunning() {
	gui.runMu.Lock()
	loopers := slices.Clone(gui.loopers)
	gui.runMu.Unlock()
	for _, ls := range loopers {
		if st := ls.Stacks[ls.Mode]; st != nil && len(st.Order) > 0 {
			ls.Stop(st.Order[len(st.Order)-1])
		}
	}
	gui.SetStopNow()
}
