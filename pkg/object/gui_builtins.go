package object

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/caokhang91/buddhist-go/pkg/tracing"
)

// GUIWindow represents a GUI window handle using pixel
type GUIWindow struct {
	ID     int64
	Window *pixelgl.Window
	Config pixelgl.WindowConfig
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
}

func (b *GUIButton) Type() ObjectType { return HASH_OBJ }
func (b *GUIButton) Inspect() string {
	return fmt.Sprintf("GUIButton{id: %d, text: %s}", b.ID, b.Text)
}

// GUITable represents a GUI table for displaying data in rows and columns
type GUITable struct {
	ID          int64
	WindowID    int64
	X           float64
	Y           float64
	Width       float64
	Height      float64
	Headers     []string
	Data        [][]string // Rows of data, each row is a slice of strings
	RowHeight   float64
	HeaderHeight float64
	OnRowClick  *Closure // Called with row index when a row is clicked
	SelectedRow int      // Currently selected row (-1 if none)
}

func (t *GUITable) Type() ObjectType { return HASH_OBJ }
func (t *GUITable) Inspect() string {
	return fmt.Sprintf("GUITable{id: %d, rows: %d, cols: %d}", t.ID, len(t.Data), len(t.Headers))
}

// Global state for GUI windows and components
var (
	guiStateMu      sync.RWMutex
	guiWindowID      int64 = 1
	guiButtonID     int64 = 1
	guiTableID      int64 = 1
	guiWindows       = make(map[int64]*GUIWindow)
	guiButtons       = make(map[int64]*GUIButton)
	guiTables        = make(map[int64]*GUITable)
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
		ID:     windowID,
		Window: nil, // Will be created when gui_run is called
		Config: windowConfig,
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

	text, _ := getStringField(config, "text", "Button")
	x, _ := getFloatField(config, "x", 0.0)
	y, _ := getFloatField(config, "y", 0.0)
	width, _ := getFloatField(config, "width", 100.0)
	height, _ := getFloatField(config, "height", 30.0)

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
		ID:       buttonID,
		WindowID: windowID,
		Text:     text,
		X:        x,
		Y:        y,
		Width:    width,
		Height:   height,
		OnClick:   onClickCallback,
	}
	guiButtons[buttonID] = button
	guiStateMu.Unlock()

	if tracing.IsEnabled() {
		tracing.Trace("Created GUI button: id=%d, text=%s, window=%d", buttonID, text, windowID)
	}

	return newStringHash(map[string]Object{
		"id":      &Integer{Value: buttonID},
		"text":    &String{Value: text},
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
		ID:          tableID,
		WindowID:    windowID,
		X:           x,
		Y:           y,
		Width:       width,
		Height:      height,
		Headers:     headers,
		Data:        data,
		RowHeight:   rowHeight,
		HeaderHeight: headerHeight,
		OnRowClick:  onRowClickCallback,
		SelectedRow: -1,
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
					
					// Clear window
					window.Window.Clear(pixel.RGB(0.95, 0.95, 0.95))
					
					// Render components
					renderWindowComponents(window)
					
					// Handle input events
					handleWindowInput(window)
				}
			}
			guiStateMu.RUnlock()

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
}

// renderButton renders a button (simple rectangle with text)
func renderButton(win *pixelgl.Window, button *GUIButton) {
	// Draw button background (simple rectangle)
	_ = pixel.R(button.X, button.Y, button.X+button.Width, button.Y+button.Height)
	// Use pixel's imdraw for simple shapes (would need to import imdraw)
	// For now, we'll just draw a simple colored rectangle using Clear with a mask
	// This is a simplified version - in a real implementation, you'd use imdraw or sprites
	
	// Note: Actual rendering would require importing pixel/imdraw or using sprites
	// This is a placeholder that shows the structure
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

// handleWindowInput handles input events for a window (mouse clicks, etc.)
func handleWindowInput(window *GUIWindow) {
	if window.Window == nil {
		return
	}

	// Check for mouse clicks
	if window.Window.JustPressed(pixelgl.MouseButtonLeft) {
		mousePos := window.Window.MousePosition()
		
		guiStateMu.RLock()
		
		// Check if any button was clicked
		for _, button := range guiButtons {
			if button.WindowID == window.ID {
				// Check if mouse is within button bounds
				if mousePos.X >= button.X && mousePos.X <= button.X+button.Width &&
					mousePos.Y >= button.Y && mousePos.Y <= button.Y+button.Height {
					// Call the onClick callback
					if button.OnClick != nil {
						callGUICallback(button.OnClick)
					}
				}
			}
		}

		// Check if any table row was clicked
		for _, table := range guiTables {
			if table.WindowID == window.ID {
				// Check if click is within table bounds
				if mousePos.X >= table.X && mousePos.X <= table.X+table.Width &&
					mousePos.Y >= table.Y && mousePos.Y <= table.Y+table.Height {
					
					// Calculate which row was clicked
					clickY := mousePos.Y - table.Y
					if clickY < table.HeaderHeight {
						// Clicked on header, ignore
						continue
					}
					
					// Calculate row index
					rowIndex := int((clickY - table.HeaderHeight) / table.RowHeight)
					if rowIndex >= 0 && rowIndex < len(table.Data) {
						table.SelectedRow = rowIndex
						
						// Call the onRowClick callback with row index
						if table.OnRowClick != nil {
							callGUICallbackWithArg(table.OnRowClick, &Integer{Value: int64(rowIndex)})
						}
					}
				}
			}
		}
		
		guiStateMu.RUnlock()
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
