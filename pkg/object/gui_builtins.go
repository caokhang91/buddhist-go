package object

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/caokhang91/buddhist-go/pkg/tracing"
)

// GUIWindow represents a GUI window handle using pixel
type GUIWindow struct {
	ID              int64
	Window          *pixelgl.Window
	Config          pixelgl.WindowConfig
	BackgroundColor [3]float64 // R,G,B 0-1; used when clearing
}

func (w *GUIWindow) Type() ObjectType { return HASH_OBJ }
func (w *GUIWindow) Inspect() string {
	if w.Window != nil {
		bounds := w.Window.Bounds()
		return fmt.Sprintf("GUIWindow{id: %d, bounds: %v}", w.ID, bounds)
	}
	return fmt.Sprintf("GUIWindow{id: %d, not created}", w.ID)
}

// GUIButton represents a GUI button (stored as metadata, actual drawing handled separately)
type GUIButton struct {
	ID       int64
	WindowID int64
	Text     string
	X        float64
	Y        float64
	Width    float64
	Height   float64
	OnClick  *Closure
	// Style: R,G,B 0-1; defaults if zero
	BgR, BgG, BgB         float64
	TextR, TextG, TextB   float64
	TextAlign string // "left", "center", "right"
}

func (b *GUIButton) Type() ObjectType { return HASH_OBJ }
func (b *GUIButton) Inspect() string {
	return fmt.Sprintf("GUIButton{id: %d, text: %s}", b.ID, b.Text)
}

// GUITable represents a GUI table for displaying data in rows and columns
type GUITable struct {
	ID           int64
	WindowID     int64
	X            float64
	Y            float64
	Width        float64
	Height       float64
	Headers      []string
	Data         [][]string // Rows of data, each row is a slice of strings
	RowHeight    float64
	HeaderHeight float64
	OnRowClick   *Closure // Called with row index when a row is clicked
	SelectedRow  int      // Currently selected row (-1 if none)
	// Style: R,G,B 0-1; used when renderTable is implemented
	HeaderBgR, HeaderBgG, HeaderBgB       float64
	HeaderTextR, HeaderTextG, HeaderTextB float64
	CellBgR, CellBgG, CellBgB             float64
	SelectedRowBgR, SelectedRowBgG, SelectedRowBgB float64
	TextR, TextG, TextB                   float64
}

func (t *GUITable) Type() ObjectType { return HASH_OBJ }
func (t *GUITable) Inspect() string {
	return fmt.Sprintf("GUITable{id: %d, rows: %d, cols: %d}", t.ID, len(t.Data), len(t.Headers))
}

// deferredGUICall is a callback (and optional arg) to run after the event loop releases guiStateMu.
// Callbacks must not run while we hold guiStateMu, else gui_alert/Lock in builtins deadlock.
type deferredGUICall struct {
	fn  *Closure
	arg Object // nil = no arg
}

// Architecture rule: the event loop holds guiStateMu.RLock() while iterating windows and
// calling handleWindowInput. We must never take guiStateMu.Lock() (or call into code that does)
// from that section. All writes (dismiss alert, etc.) and all VM callbacks are queued and
// run only after RUnlock, in order: first deferred state updates, then deferred callbacks.

var (
	guiStateMu              sync.RWMutex
	guiCallbackQueue        []deferredGUICall
	guiDeferredStateUpdates []func() // e.g. dismiss alert; run after RUnlock, before callbacks
	guiCallbackMu           sync.Mutex
	guiWindowID      int64 = 1
	guiButtonID      int64 = 1
	guiTableID       int64 = 1
	guiWindows       = make(map[int64]*GUIWindow)
	guiButtons       = make(map[int64]*GUIButton)
	guiTables        = make(map[int64]*GUITable)
	guiAlerts        = make(map[int64]string) // windowID -> message; shown until OK clicked
	guiRunStarted    bool
	guiRunChan       chan struct{}
	guiEventLoop     func()
)

func init() {
	guiRunChan = make(chan struct{})
}

// guiWindowBuiltin creates a new GUI window configuration
// Usage: gui_window({"title": "My App", "width": 800, "height": 600, "vsync": true})
func guiWindowBuiltin(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}

	config, ok := args[0].(*Hash)
	if !ok {
		return newError("argument to `gui_window` must be HASH, got %s", args[0].Type())
	}

	// Extract window configuration
	title, _ := getStringField(config, "title", "Window")
	width, _ := getIntField(config, "width", 800)
	height, _ := getIntField(config, "height", 600)
	vsync, _ := getBoolField(config, "vsync", true)
	br, bg, bb := getColorField(config, "backgroundColor", 0.95, 0.95, 0.95)

	guiStateMu.Lock()
	windowID := guiWindowID
	guiWindowID++

	// Create window config (window will be created when gui_run is called)
	windowConfig := pixelgl.WindowConfig{
		Title:  title,
		Bounds: pixel.R(0, 0, float64(width), float64(height)),
		VSync:  vsync,
	}

	window := &GUIWindow{
		ID:              windowID,
		Window:          nil, // Will be created when gui_run is called
		Config:          windowConfig,
		BackgroundColor:  [3]float64{br, bg, bb},
	}
	guiWindows[windowID] = window
	guiStateMu.Unlock()

	if tracing.IsEnabled() {
		tracing.Trace("Created GUI window config: id=%d, title=%s, size=%dx%d", windowID, title, width, height)
	}

	// Return window as a hash
	return newStringHash(map[string]Object{
		"id":     &Integer{Value: windowID},
		"title":  &String{Value: title},
		"width":  &Integer{Value: int64(width)},
		"height": &Integer{Value: int64(height)},
		"vsync":  &Boolean{Value: vsync},
		"_type":  &String{Value: "GUIWindow"},
	})
}

// guiButtonBuiltin creates a new GUI button metadata
// Usage: gui_button(window, {"text": "Click Me", "x": 100, "y": 100, "width": 200, "height": 50, "onClick": fn() { ... }})
func guiButtonBuiltin(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments. got=%d, want=2", len(args))
	}

	// First argument should be the window
	windowHash, ok := args[0].(*Hash)
	if !ok {
		return newError("first argument to `gui_button` must be HASH (window), got %s", args[0].Type())
	}

	windowIDObj, ok := getHashValue(windowHash, "id")
	if !ok {
		return newError("window must have an 'id' field")
	}

	windowID, errObj := intFromObject(windowIDObj, "gui_button", "window.id")
	if errObj != nil {
		return errObj
	}

	// Check if window exists
	guiStateMu.RLock()
	_, exists := guiWindows[windowID]
	guiStateMu.RUnlock()
	if !exists {
		return newError("window with id %d does not exist", windowID)
	}

	// Second argument should be button configuration
	config, ok := args[1].(*Hash)
	if !ok {
		return newError("second argument to `gui_button` must be HASH, got %s", args[1].Type())
	}

	textVal, _ := getStringField(config, "text", "Button")
	x, _ := getFloatField(config, "x", 0.0)
	y, _ := getFloatField(config, "y", 0.0)
	width, _ := getFloatField(config, "width", 100.0)
	height, _ := getFloatField(config, "height", 30.0)
	bgR, bgG, bgB := getColorField(config, "bgColor", 0.85, 0.85, 0.9)
	textR, textG, textB := getColorField(config, "textColor", 0.1, 0.1, 0.1)
	textAlign, _ := getStringField(config, "textAlign", "left")
	if textAlign != "left" && textAlign != "center" && textAlign != "right" {
		textAlign = "left"
	}

	// Check for onClick callback
	var onClickCallback *Closure
	if onClickObj, ok := getHashValue(config, "onClick"); ok {
		if closure, ok := onClickObj.(*Closure); ok {
			onClickCallback = closure
		} else {
			return newError("`gui_button` onClick must be FUNCTION, got %s", onClickObj.Type())
		}
	}

	guiStateMu.Lock()
	buttonID := guiButtonID
	guiButtonID++
	button := &GUIButton{
		ID:        buttonID,
		WindowID:  windowID,
		Text:      textVal,
		X:         x,
		Y:         y,
		Width:     width,
		Height:    height,
		OnClick:   onClickCallback,
		BgR:       bgR,
		BgG:       bgG,
		BgB:       bgB,
		TextR:      textR,
		TextG:      textG,
		TextB:      textB,
		TextAlign:  textAlign,
	}
	guiButtons[buttonID] = button
	guiStateMu.Unlock()

	if tracing.IsEnabled() {
		tracing.Trace("Created GUI button: id=%d, text=%s, window=%d", buttonID, textVal, windowID)
	}

	return newStringHash(map[string]Object{
		"id":      &Integer{Value: buttonID},
		"text":    &String{Value: textVal},
		"x":       &Float{Value: x},
		"y":       &Float{Value: y},
		"width":   &Float{Value: width},
		"height":  &Float{Value: height},
		"window":  &Integer{Value: windowID},
		"_type":   &String{Value: "GUIButton"},
	})
}

// guiShowBuiltin marks a window to be shown (windows are shown when gui_run is called)
// Usage: gui_show(window)
func guiShowBuiltin(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}

	windowHash, ok := args[0].(*Hash)
	if !ok {
		return newError("argument to `gui_show` must be HASH (window), got %s", args[0].Type())
	}

	windowIDObj, ok := getHashValue(windowHash, "id")
	if !ok {
		return newError("window must have an 'id' field")
	}

	windowID, errObj := intFromObject(windowIDObj, "gui_show", "window.id")
	if errObj != nil {
		return errObj
	}

	guiStateMu.RLock()
	_, exists := guiWindows[windowID]
	guiStateMu.RUnlock()

	if !exists {
		return newError("window with id %d does not exist", windowID)
	}

	if tracing.IsEnabled() {
		tracing.Trace("Window %d will be shown when gui_run is called", windowID)
	}

	return &Null{}
}

// guiCloseBuiltin marks a window for closing
// Usage: gui_close(window)
func guiCloseBuiltin(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}

	windowHash, ok := args[0].(*Hash)
	if !ok {
		return newError("argument to `gui_close` must be HASH (window), got %s", args[0].Type())
	}

	windowIDObj, ok := getHashValue(windowHash, "id")
	if !ok {
		return newError("window must have an 'id' field")
	}

	windowID, errObj := intFromObject(windowIDObj, "gui_close", "window.id")
	if errObj != nil {
		return errObj
	}

	guiStateMu.Lock()
	window, exists := guiWindows[windowID]
	if exists {
		// Close the window if it's been created
		if window.Window != nil && !window.Window.Closed() {
			window.Window.SetClosed(true)
		}
		delete(guiWindows, windowID)
		delete(guiAlerts, windowID)
		// Also remove buttons and tables for this window
		for id, btn := range guiButtons {
			if btn.WindowID == windowID {
				delete(guiButtons, id)
			}
		}
		for id, tbl := range guiTables {
			if tbl.WindowID == windowID {
				delete(guiTables, id)
			}
		}
	}
	guiStateMu.Unlock()

	if !exists {
		return newError("window with id %d does not exist", windowID)
	}

	if tracing.IsEnabled() {
		tracing.Trace("Closed GUI window: id=%d", windowID)
	}
	return &Null{}
}

// guiAlertBuiltin shows a modal alert (message + OK) on the given window.
// Usage: gui_alert(window, "Message here")
func guiAlertBuiltin(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments. got=%d, want=2", len(args))
	}
	windowHash, ok := args[0].(*Hash)
	if !ok {
		return newError("first argument to `gui_alert` must be HASH (window), got %s", args[0].Type())
	}
	windowIDObj, ok := getHashValue(windowHash, "id")
	if !ok {
		return newError("window must have an 'id' field")
	}
	windowID, errObj := intFromObject(windowIDObj, "gui_alert", "window.id")
	if errObj != nil {
		return errObj
	}
	guiStateMu.RLock()
	_, exists := guiWindows[windowID]
	guiStateMu.RUnlock()
	if !exists {
		return newError("window with id %d does not exist", windowID)
	}
	msg := args[1].Inspect()
	if s, ok := args[1].(*String); ok {
		msg = s.Value
	}
	guiStateMu.Lock()
	guiAlerts[windowID] = msg
	guiStateMu.Unlock()
	return &Null{}
}

// guiTableBuiltin creates a new GUI table for displaying data
// Usage: gui_table(window, {"x": 50, "y": 400, "width": 700, "height": 300, "headers": ["Name", "Address", "City"], "data": [[...], [...]], "onRowClick": fn(row) { ... }})
func guiTableBuiltin(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments. got=%d, want=2", len(args))
	}

	// First argument should be the window
	windowHash, ok := args[0].(*Hash)
	if !ok {
		return newError("first argument to `gui_table` must be HASH (window), got %s", args[0].Type())
	}

	windowIDObj, ok := getHashValue(windowHash, "id")
	if !ok {
		return newError("window must have an 'id' field")
	}

	windowID, errObj := intFromObject(windowIDObj, "gui_table", "window.id")
	if errObj != nil {
		return errObj
	}

	// Check if window exists
	guiStateMu.RLock()
	_, exists := guiWindows[windowID]
	guiStateMu.RUnlock()
	if !exists {
		return newError("window with id %d does not exist", windowID)
	}

	// Second argument should be table configuration
	config, ok := args[1].(*Hash)
	if !ok {
		return newError("second argument to `gui_table` must be HASH, got %s", args[1].Type())
	}

	x, _ := getFloatField(config, "x", 0.0)
	y, _ := getFloatField(config, "y", 0.0)
	width, _ := getFloatField(config, "width", 500.0)
	height, _ := getFloatField(config, "height", 300.0)
	rowHeight, _ := getFloatField(config, "rowHeight", 25.0)
	headerHeight, _ := getFloatField(config, "headerHeight", 30.0)
	headerBgR, headerBgG, headerBgB := getColorField(config, "headerBg", 0.6, 0.6, 0.65)
	headerTextR, headerTextG, headerTextB := getColorField(config, "headerTextColor", 0.1, 0.1, 0.1)
	cellBgR, cellBgG, cellBgB := getColorField(config, "cellBg", 0.98, 0.98, 0.98)
	selR, selG, selB := getColorField(config, "selectedRowBg", 0.7, 0.8, 0.95)
	textR, textG, textB := getColorField(config, "textColor", 0.1, 0.1, 0.1)

	// Extract headers
	var headers []string
	if headersObj, ok := getHashValue(config, "headers"); ok {
		if arr, ok := headersObj.(*Array); ok {
			headers = make([]string, 0, len(arr.Elements))
			for _, elem := range arr.Elements {
				if str, ok := elem.(*String); ok {
					headers = append(headers, str.Value)
				}
			}
		}
	}

	// Extract data rows
	var data [][]string
	if dataObj, ok := getHashValue(config, "data"); ok {
		if arr, ok := dataObj.(*Array); ok {
			data = make([][]string, 0, len(arr.Elements))
			for _, rowObj := range arr.Elements {
				if rowArr, ok := rowObj.(*Array); ok {
					row := make([]string, 0, len(rowArr.Elements))
					for _, cellObj := range rowArr.Elements {
						cellStr := cellObj.Inspect()
						// Remove quotes if it's a string
						if len(cellStr) >= 2 && cellStr[0] == '"' && cellStr[len(cellStr)-1] == '"' {
							cellStr = cellStr[1 : len(cellStr)-1]
						}
						row = append(row, cellStr)
					}
					data = append(data, row)
				}
			}
		}
	}

	// Check for onRowClick callback
	var onRowClickCallback *Closure
	if onRowClickObj, ok := getHashValue(config, "onRowClick"); ok {
		if closure, ok := onRowClickObj.(*Closure); ok {
			onRowClickCallback = closure
		} else {
			return newError("`gui_table` onRowClick must be FUNCTION, got %s", onRowClickObj.Type())
		}
	}

	guiStateMu.Lock()
	tableID := guiTableID
	guiTableID++
	table := &GUITable{
		ID:               tableID,
		WindowID:         windowID,
		X:                x,
		Y:                y,
		Width:            width,
		Height:           height,
		Headers:          headers,
		Data:             data,
		RowHeight:        rowHeight,
		HeaderHeight:     headerHeight,
		OnRowClick:       onRowClickCallback,
		SelectedRow:      -1,
		HeaderBgR:        headerBgR,
		HeaderBgG:        headerBgG,
		HeaderBgB:        headerBgB,
		HeaderTextR:      headerTextR,
		HeaderTextG:      headerTextG,
		HeaderTextB:      headerTextB,
		CellBgR:          cellBgR,
		CellBgG:          cellBgG,
		CellBgB:          cellBgB,
		SelectedRowBgR:   selR,
		SelectedRowBgG:   selG,
		SelectedRowBgB:   selB,
		TextR:            textR,
		TextG:            textG,
		TextB:            textB,
	}
	guiTables[tableID] = table
	guiStateMu.Unlock()

	if tracing.IsEnabled() {
		tracing.Trace("Created GUI table: id=%d, rows=%d, cols=%d, window=%d", tableID, len(data), len(headers), windowID)
	}

	// Return table as a hash
	return newStringHash(map[string]Object{
		"id":          &Integer{Value: tableID},
		"x":           &Float{Value: x},
		"y":           &Float{Value: y},
		"width":       &Float{Value: width},
		"height":      &Float{Value: height},
		"window":      &Integer{Value: windowID},
		"_type":       &String{Value: "GUITable"},
	})
}

// guiRunBuiltin runs the GUI event loop using pixelgl
// Usage: gui_run()
// Note: This must be called to actually create and display windows
func guiRunBuiltin(args ...Object) Object {
	if len(args) != 0 {
		return newError("wrong number of arguments. got=%d, want=0", len(args))
	}

	guiStateMu.Lock()
	if guiRunStarted {
		guiStateMu.Unlock()
		return newError("GUI event loop is already running")
	}
	guiRunStarted = true
	guiStateMu.Unlock()

	// Create the event loop function
	eventLoop := func() {
		// Create all windows
		guiStateMu.Lock()
		windowsToCreate := make([]*GUIWindow, 0, len(guiWindows))
		for _, window := range guiWindows {
			if window.Window == nil {
				windowsToCreate = append(windowsToCreate, window)
			}
		}
		guiStateMu.Unlock()

		// Create windows (must be on main thread)
		for _, window := range windowsToCreate {
			win, err := pixelgl.NewWindow(window.Config)
			if err != nil {
				if tracing.IsEnabled() {
					tracing.Trace("Failed to create window %d: %v", window.ID, err)
				}
				continue
			}
			window.Window = win
		}

		// Main event loop
		for {
			allClosed := true
			
			guiStateMu.RLock()
			for _, window := range guiWindows {
				if window.Window == nil {
					continue
				}
				if !window.Window.Closed() {
					allClosed = false
					window.Window.Update()

					// Clear window with optional backgroundColor from config
					window.Window.Clear(pixel.RGB(window.BackgroundColor[0], window.BackgroundColor[1], window.BackgroundColor[2]))

					// Render components
					renderWindowComponents(window)
					
					// Handle input events
					handleWindowInput(window)
				}
			}
			guiStateMu.RUnlock()

			// Run deferred work only when not holding guiStateMu (avoids deadlock).
			// Order: state updates (e.g. dismiss alert) then VM callbacks.
			runDeferredStateUpdates()
			runDeferredCallbacks()

			if allClosed {
				break
			}
		}
	}

	// Run pixelgl on the current (main) thread — blocks until all windows are closed.
	// When using ./buddhist-go -g <file>, we're on main goroutine so the window stays open.
	runtime.LockOSThread()
	pixelgl.Run(eventLoop)

	guiStateMu.Lock()
	guiRunStarted = false
	guiStateMu.Unlock()

	if tracing.IsEnabled() {
		tracing.Trace("GUI event loop ended")
	}

	return &Null{}
}

// renderWindowComponents renders all GUI components for a window
func renderWindowComponents(window *GUIWindow) {
	if window.Window == nil {
		return
	}

	guiStateMu.RLock()
	defer guiStateMu.RUnlock()

	// Render buttons (simple text-based rendering)
	for _, button := range guiButtons {
		if button.WindowID == window.ID {
			renderButton(window.Window, button)
		}
	}

	// Render tables
	for _, table := range guiTables {
		if table.WindowID == window.ID {
			renderTable(window.Window, table)
		}
	}
	// Render alert overlay if active
	if msg, ok := guiAlerts[window.ID]; ok && msg != "" {
		renderAlert(window.Window, msg)
	}
}

// renderAlert draws a modal overlay: dim background, centered box with message + OK button.
func renderAlert(win *pixelgl.Window, message string) {
	winW := win.Bounds().W()
	winH := win.Bounds().H()
	const boxW, boxH = 280.0, 120.0
	boxMinX := winW*0.5 - boxW*0.5
	boxMinY := winH*0.5 - boxH*0.5
	boxMaxX := winW*0.5 + boxW*0.5
	boxMaxY := winH*0.5 + boxH*0.5

	imd := imdraw.New(nil)
	// Overlay (dim)
	imd.Color = pixel.RGB(0.2, 0.2, 0.22)
	imd.Push(pixel.V(0, 0), pixel.V(winW, winH))
	imd.Rectangle(0)
	imd.Draw(win)
	// Box
	imd.Color = pixel.RGB(0.97, 0.97, 0.98)
	imd.Push(pixel.V(boxMinX, boxMinY), pixel.V(boxMaxX, boxMaxY))
	imd.Rectangle(0)
	imd.Draw(win)
	// OK button
	const okW, okH = 80.0, 28.0
	okMinX := winW*0.5 - okW*0.5
	okMinY := boxMinY + 10.0
	imd.Color = pixel.RGB(0.85, 0.85, 0.9)
	imd.Push(pixel.V(okMinX, okMinY), pixel.V(okMinX+okW, okMinY+okH))
	imd.Rectangle(0)
	imd.Draw(win)

	if text.Atlas7x13 != nil {
		const pad = 12.0
		// Message (truncate if too long)
		msg := message
		if len(msg) > 42 {
			msg = msg[:39] + "..."
		}
		txt := text.New(pixel.V(boxMinX+pad, boxMaxY-pad), text.Atlas7x13)
		txt.Color = pixel.RGB(0.1, 0.1, 0.1)
		txt.WriteString(msg)
		txt.Draw(win, pixel.IM)
		// OK button label (centered in button area)
		okY := boxMinY + 16.0
		measure := text.New(pixel.ZV, text.Atlas7x13)
		okW := measure.BoundsOf("OK").W()
		okTxt := text.New(pixel.V(winW*0.5-okW*0.5, okY), text.Atlas7x13)
		okTxt.Color = pixel.RGB(0.1, 0.1, 0.1)
		okTxt.WriteString("OK")
		okTxt.Draw(win, pixel.IM)
	}
}

// renderButton renders a button (rectangle background + text label).
// Script (button.X, button.Y) is top-left; converted to pixel (0,0=bottom-left) via winHeight.
func renderButton(win *pixelgl.Window, button *GUIButton) {
	const pad = 8.0
	winHeight := win.Bounds().H()
	_, pixelY := scriptTopLeftToPixel(button.X, button.Y, button.Height, winHeight)

	// 1. Draw button background (filled rectangle)
	imd := imdraw.New(nil)
	imd.Color = pixel.RGB(button.BgR, button.BgG, button.BgB)
	min := pixel.V(button.X, pixelY)
	max := pixel.V(button.X+button.Width, pixelY+button.Height)
	imd.Push(min, max)
	imd.Rectangle(0)
	imd.Draw(win)

	// 2. Draw button text with alignment (left/center/right)
	if button.Text != "" && text.Atlas7x13 != nil {
		txtOriginY := winHeight - button.Y - pad
		var originX float64
		measure := text.New(pixel.ZV, text.Atlas7x13)
		textW := measure.BoundsOf(button.Text).W()
		switch button.TextAlign {
		case "center":
			originX = button.X + (button.Width-textW)*0.5
		case "right":
			originX = button.X + button.Width - pad - textW
		default: // "left"
			originX = button.X + pad
		}
		txt := text.New(pixel.V(originX, txtOriginY), text.Atlas7x13)
		txt.Color = pixel.RGB(button.TextR, button.TextG, button.TextB)
		txt.WriteString(button.Text)
		txt.Draw(win, pixel.IM)
	}
}

// renderTable renders a table with headers and data rows
func renderTable(win *pixelgl.Window, table *GUITable) {
	// Note: Actual rendering would require text rendering using pixel/text
	// This is a placeholder structure showing how it would work
	// In a real implementation, you would:
	// 1. Draw header background
	// 2. Draw header text for each column
	// 3. Draw row backgrounds (alternating colors, highlight selected)
	// 4. Draw cell text for each row
	// 5. Draw grid lines
}

// handleWindowInput handles input events for a window (mouse clicks, etc.).
// Mouse position is in pixel (0,0=bottom-left); converted to script (0,0=top-left) for hit-test.
func handleWindowInput(window *GUIWindow) {
	if window.Window == nil {
		return
	}

	if window.Window.JustPressed(pixelgl.MouseButtonLeft) {
		mousePos := window.Window.MousePosition()
		winW := window.Window.Bounds().W()
		winH := window.Window.Bounds().H()
		scriptY := winH - mousePos.Y

		guiStateMu.RLock()
		alertMsg := guiAlerts[window.ID]
		guiStateMu.RUnlock()
		if alertMsg != "" {
			// Alert is shown: only OK button is clickable (box 280x120, OK 80x28 centered at bottom).
			// Do not Lock here — we're still under event loop's RLock. Queue dismiss to run after RUnlock.
			const boxH = 120.0
			boxMinY := winH*0.5 - boxH*0.5
			okMinX := winW*0.5 - 40.0
			if mousePos.X >= okMinX && mousePos.X <= okMinX+80 &&
				mousePos.Y >= boxMinY+10 && mousePos.Y <= boxMinY+38 {
				queueDismissAlert(window.ID)
			}
			return
		}

		// Collect callback (and table arg) under lock, then release and invoke — avoids deadlock
		// when callback calls gui_alert() or other builtins that take guiStateMu.Lock().
		var onClick *Closure
		var onRowClick *Closure
		var rowClickArg *Integer
		var selectedTable *GUITable
		var selectedRow int
		guiStateMu.RLock()
		for _, button := range guiButtons {
			if button.WindowID == window.ID {
				if mousePos.X >= button.X && mousePos.X <= button.X+button.Width &&
					scriptY >= button.Y && scriptY <= button.Y+button.Height {
					onClick = button.OnClick
					break
				}
			}
		}
		if onClick == nil {
			for _, tbl := range guiTables {
				if tbl.WindowID == window.ID {
					if mousePos.X >= tbl.X && mousePos.X <= tbl.X+tbl.Width &&
						scriptY >= tbl.Y && scriptY <= tbl.Y+tbl.Height {
						localY := scriptY - tbl.Y
						if localY >= tbl.HeaderHeight {
							rowIndex := int((localY - tbl.HeaderHeight) / tbl.RowHeight)
							if rowIndex >= 0 && rowIndex < len(tbl.Data) {
								selectedTable = tbl
								selectedRow = rowIndex
								onRowClick = tbl.OnRowClick
								rowClickArg = &Integer{Value: int64(rowIndex)}
								break
							}
						}
					}
				}
			}
		}
		guiStateMu.RUnlock()

		if selectedTable != nil {
			selectedTable.SelectedRow = selectedRow
		}
		// Queue callbacks to run after guiStateMu is released (avoids deadlock when
		// callback calls gui_alert or other builtins that take Lock).
		if onClick != nil {
			queueGUICallback(onClick, nil)
		} else if onRowClick != nil && rowClickArg != nil {
			queueGUICallback(onRowClick, rowClickArg)
		}
	}
}

// queueDismissAlert queues removal of the alert for windowID. Must be called from handleWindowInput
// (which runs under RLock); the actual delete runs in runDeferredStateUpdates() after RUnlock.
func queueDismissAlert(windowID int64) {
	guiCallbackMu.Lock()
	guiDeferredStateUpdates = append(guiDeferredStateUpdates, func() {
		guiStateMu.Lock()
		delete(guiAlerts, windowID)
		guiStateMu.Unlock()
	})
	guiCallbackMu.Unlock()
}

// runDeferredStateUpdates runs all queued state updates (e.g. dismiss alert). Must be called when not holding guiStateMu.
func runDeferredStateUpdates() {
	guiCallbackMu.Lock()
	updates := guiDeferredStateUpdates
	guiDeferredStateUpdates = nil
	guiCallbackMu.Unlock()
	for _, fn := range updates {
		fn()
	}
}

// queueGUICallback queues a callback to run when runDeferredCallbacks() is called (after RUnlock).
func queueGUICallback(fn *Closure, arg Object) {
	if fn == nil {
		return
	}
	guiCallbackMu.Lock()
	guiCallbackQueue = append(guiCallbackQueue, deferredGUICall{fn: fn, arg: arg})
	guiCallbackMu.Unlock()
}

// runDeferredCallbacks drains the queue and runs each callback. Must be called when not holding guiStateMu.
func runDeferredCallbacks() {
	guiCallbackMu.Lock()
	q := guiCallbackQueue
	guiCallbackQueue = nil
	guiCallbackMu.Unlock()
	for _, c := range q {
		if c.arg != nil {
			callGUICallbackWithArg(c.fn, c.arg)
		} else {
			callGUICallback(c.fn)
		}
	}
}

// Helper functions for extracting fields from Hash objects

func getStringField(hash *Hash, key string, defaultValue string) (string, bool) {
	value, ok := getHashValue(hash, key)
	if !ok {
		return defaultValue, false
	}
	if str, ok := value.(*String); ok {
		return str.Value, true
	}
	return defaultValue, false
}

func getIntField(hash *Hash, key string, defaultValue int) (int, bool) {
	value, ok := getHashValue(hash, key)
	if !ok {
		return defaultValue, false
	}
	if intObj, ok := value.(*Integer); ok {
		return int(intObj.Value), true
	}
	if floatObj, ok := value.(*Float); ok {
		return int(floatObj.Value), true
	}
	return defaultValue, false
}

func getFloatField(hash *Hash, key string, defaultValue float64) (float64, bool) {
	value, ok := getHashValue(hash, key)
	if !ok {
		return defaultValue, false
	}
	if floatObj, ok := value.(*Float); ok {
		return floatObj.Value, true
	}
	if intObj, ok := value.(*Integer); ok {
		return float64(intObj.Value), true
	}
	return defaultValue, false
}

func getBoolField(hash *Hash, key string, defaultValue bool) (bool, bool) {
	value, ok := getHashValue(hash, key)
	if !ok {
		return defaultValue, false
	}
	if boolObj, ok := value.(*Boolean); ok {
		return boolObj.Value, true
	}
	return defaultValue, false
}

// getColorField reads an optional color from hash[key]. Value must be a hash with "r","g","b"
// (Float 0-1 or Integer 0-255). Returns (r,g,b) in 0-1 for pixel.RGB. Uses defaults if missing.
func getColorField(hash *Hash, key string, defaultR, defaultG, defaultB float64) (r, g, b float64) {
	value, ok := getHashValue(hash, key)
	if !ok {
		return defaultR, defaultG, defaultB
	}
	colorHash, ok := value.(*Hash)
	if !ok {
		return defaultR, defaultG, defaultB
	}
	r, _ = getFloatField(colorHash, "r", defaultR)
	g, _ = getFloatField(colorHash, "g", defaultG)
	b, _ = getFloatField(colorHash, "b", defaultB)
	// Normalize: if > 1 assume 0-255
	if r > 1 {
		r = r / 255
	}
	if g > 1 {
		g = g / 255
	}
	if b > 1 {
		b = b / 255
	}
	return r, g, b
}

// scriptTopLeftToPixel converts script coords (x,y)=top-left of widget to pixel coords
// (0,0)=bottom-left. pixelY is the bottom-edge Y of the widget.
func scriptTopLeftToPixel(scriptX, scriptY, widgetHeight, winHeight float64) (pixelX, pixelY float64) {
	return scriptX, winHeight - scriptY - widgetHeight
}

// callGUICallback calls a GUI event callback closure
func callGUICallback(callback *Closure) {
	if callback == nil {
		return
	}

	closureCallerMu.RLock()
	caller := closureCaller
	closureCallerMu.RUnlock()

	if caller == nil {
		return
	}

	// Call the closure with no arguments (can be extended to pass event data)
	_, err := caller(callback)
	if err != nil {
		if tracing.IsEnabled() {
			tracing.Trace("GUI callback error: %v", err)
		}
	}
}

// callGUICallbackWithArg calls a GUI event callback closure with one argument
func callGUICallbackWithArg(callback *Closure, arg Object) {
	if callback == nil {
		return
	}

	closureCallerMu.RLock()
	caller := closureCaller
	closureCallerMu.RUnlock()

	if caller == nil {
		return
	}

	// Create a temporary closure that includes the argument
	// This is a simplified approach - in practice, you'd need to properly handle arguments
	// For now, we'll store the argument in the closure's free variables
	// Note: This requires modifying how closures work, so for now we'll just call with no args
	// A proper implementation would need to pass arguments through the VM
	
	// Temporary workaround: just call without args for now
	// TODO: Implement proper argument passing for GUI callbacks
	_, err := caller(callback)
	if err != nil {
		if tracing.IsEnabled() {
			tracing.Trace("GUI callback error: %v", err)
		}
	}
}
