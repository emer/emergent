// Copyright (c) 2026, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package egui

import (
	"sync/atomic"
	"testing"
	"time"

	"cogentcore.org/core/tree"
	"github.com/stretchr/testify/assert"
)

// destroySim records whether it is destroyed, and whether its GUI was
// still running then.
type destroySim struct {
	gui              *GUI
	destroyed        atomic.Bool
	destroyedRunning atomic.Bool
}

func (ds *destroySim) Destroy() {
	ds.destroyedRunning.Store(ds.gui.IsRunning())
	ds.destroyed.Store(true)
}

// TestSimFrameDestroy checks that destroying the frame of a running sim
// stops it, and destroys the sim only once it has stopped, as the next
// sim must not be made before.
func TestSimFrameDestroy(t *testing.T) {
	gui := &GUI{}
	ds := &destroySim{gui: gui}
	sf := tree.New[simFrame]()
	sf.gui, sf.sim, sf.released = gui, ds, make(chan struct{})

	// a run that stops a while after it is told to
	gui.StartRun()
	go func() {
		for !gui.StopNow() {
			time.Sleep(time.Millisecond)
		}
		time.Sleep(50 * time.Millisecond)
		gui.Stopped(nil, nil)
	}()

	released := sf.released
	sf.Destroy()
	assert.False(t, ds.destroyed.Load(), "destroyed while running")
	assert.Len(t, pendingReleases(), 1)
	select {
	case <-released:
	case <-time.After(releaseWait):
		t.Fatal("not released")
	}
	assert.True(t, ds.destroyed.Load())
	assert.False(t, ds.destroyedRunning.Load(), "destroyed before it stopped")
	time.Sleep(10 * time.Millisecond) // the release is removed after it is closed
	assert.Empty(t, pendingReleases())

	// a sim that is not running is destroyed right away
	ds2 := &destroySim{gui: &GUI{}}
	sf2 := tree.New[simFrame]()
	sf2.gui, sf2.sim, sf2.released = ds2.gui, ds2, make(chan struct{})
	sf2.Destroy()
	assert.True(t, ds2.destroyed.Load())
	assert.Empty(t, pendingReleases())
}
