//go:build !windows

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
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"
)

// watchResize keeps the terminal's dimensions in sync with the TTY. On each
// SIGWINCH it re-queries the size and updates the terminal so cursor and
// repaint maths stay correct after the window is resized. The returned function
// stops the watcher and must be called when the explorer exits.
func watchResize(fileDescriptor int, terminal *term.Terminal) func() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGWINCH)

	done := make(chan struct{})

	go func() {
		for {
			select {
			case <-signals:
				if width, height, err := term.GetSize(fileDescriptor); err == nil {
					_ = terminal.SetSize(width, height)
				}
			case <-done:
				return
			}
		}
	}()

	return func() {
		signal.Stop(signals)
		close(done)
	}
}
