# GUI Roadmap for Buddhist-Go

**Status**: Experimental / Pre-alpha (as of January 2026)  
**Goal**: Provide a simple, native-feeling GUI toolkit for Buddhist-Go scripts so users can build desktop-like applications (tables, buttons, forms, basic drawing) without leaving the language.  
**Philosophy**: Keep it lightweight, script-friendly, and Go-native. No heavy frameworks. Prefer immediate-mode or minimal retained-mode APIs. Support macOS (Apple Silicon), Windows, Linux.

## Core Principles
- GUI calls must be simple and declarative in .bl scripts (like your `gui_window`, `gui_table`, `gui_button` example).
- Callbacks (`onClick`, `onRowClick`) map to Buddhist-Go functions.
- Output to terminal when GUI is not available (headless/SSH/remote).
- Avoid blocking the main thread (use goroutines under the hood).
- No external package manager dependency for end-users (bundle or use stdlib + minimal deps).

## Current State (Jan 2026)
- Private/local prototype exists (GLFW + OpenGL backend).
- Example script: `examples/address_management.bl` (table + buttons + row selection).
- Known issues:
  - macOS Apple Silicon: "Cocoa: Failed to find service port for display" when run via SSH/headless/remote.
  - No fallback mode yet.
  - Only basic widgets (window, table, button).
  - No layout system (absolute positioning only).
  - No persistence (data reset on close).

## Phase 1 – Minimal Viable GUI (Q1 2026 – 2–4 weeks effort)
- **Backend choice**: Switch from GLFW to **Gio UI** (gioui.org) or **webview** (github.com/webview/webview_go)  
  - Gio: pure Go, Metal on macOS, good performance, no Cocoa port issues.  
  - webview: easiest cross-platform, embed browser, render HTML + JS bridge → avoids native display errors entirely.
- **Core functions**:
  - `gui_window(options_map)` → create & return window handle
  - `gui_show(window)`
  - `gui_run()` → start event loop (non-blocking if possible)
  - `gui_button(window, options_map)` → text, position, onClick callback
  - `gui_table(window, options_map)` → headers, data (array of arrays), onRowClick
- **Fallback**: If no display → print HTML representation to stdout + suggestion to open in browser
- **Flag support**: `--gui`, `-g`, or env `BUDDHIST_GUI=1`
- **Deliverables**:
  - Update README with GUI section + example
  - Fix Cocoa error by switching backend
  - Commit `examples/address_management.bl`

## Phase 2 – Usable Widgets & Layout (Q2 2026 – 4–8 weeks)
- Add:
  - `gui_label`, `gui_text_input`, `gui_checkbox`, `gui_dropdown`
  - Basic layout: `gui_column`, `gui_row`, `gui_box` (flex-like)
  - Dialogs: `gui_alert`, `gui_confirm`, `gui_input_dialog`
  - Events: `onClose`, `onResize`, `onKey`
- Data binding: two-way sync between table/input and Buddhist-Go variables
- Styling: basic theme (colors, fonts) via options_map
- Persistence: optional save/load window state (position/size)

## Phase 3 – Polish & Advanced (Q3–Q4 2026)
- Canvas/drawing: `gui_canvas`, draw lines/rects/text/images
- File dialog: open/save file picker
- Menu bar, tabs, scroll views
- Cross-platform testing (macOS, Windows, Linux)
- Packaging: bundle interpreter + GUI deps into single binary (via `go build -ldflags` or tinygo?)
- Documentation: separate `GUI.md` or tutorial video

## Known Challenges & Decisions
- **macOS Apple Silicon**: GLFW causes Cocoa service port errors in headless/SSH. → Prefer Gio or webview.
- **Performance**: Table with 1000+ rows → use virtual scrolling (Gio supports it natively).
- **Concurrency**: GUI loop must not block VM → run in separate goroutine, use channels for callback communication.
- **Dependencies**: Keep minimal (Gio or webview only) – no npm/electron-style bloat.

## How to Test Current Prototype (local only)
```bash
# From local Terminal.app (not SSH!)
./buddhist-go --gui examples/address_management.bl