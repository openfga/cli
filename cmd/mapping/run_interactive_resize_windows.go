//go:build windows

/*
Copyright © 2023 OpenFGA

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package mapping

import (
	"time"

	"golang.org/x/term"
)

// resizePollInterval is how often the console size is re-queried. Windows has no
// SIGWINCH, so a resize is detected by polling rather than a signal.
const resizePollInterval = 250 * time.Millisecond

// watchResize keeps the terminal's dimensions in sync with the console. Windows
// has no SIGWINCH, so the size is polled: on each tick the current console size
// is queried and, when it has changed, pushed to the terminal so cursor and
// repaint maths stay correct after the window is resized. term.Terminal
// serialises SetSize against an in-progress ReadLine with an internal lock, so
// the repaint applies on the next keystroke — the same behaviour as the
// signal-driven Unix path. The returned function stops the watcher and must be
// called when the explorer exits.
func watchResize(fileDescriptor int, terminal *term.Terminal) func() {
	done := make(chan struct{})

	go func() {
		ticker := time.NewTicker(resizePollInterval)
		defer ticker.Stop()

		// Seed from the size runRaw already applied so an unchanged console does
		// not trigger a redundant SetSize.
		lastWidth, lastHeight, _ := term.GetSize(fileDescriptor)

		for {
			select {
			case <-ticker.C:
				width, height, err := term.GetSize(fileDescriptor)
				if err != nil || (width == lastWidth && height == lastHeight) {
					continue
				}

				lastWidth, lastHeight = width, height
				_ = terminal.SetSize(width, height)
			case <-done:
				return
			}
		}
	}()

	return func() { close(done) }
}
