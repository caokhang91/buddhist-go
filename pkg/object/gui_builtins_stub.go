//go:build !gui

package object

const guiUnavailableMsg = "GUI is not available in this build; build with -tags gui to enable (requires CGO and OpenGL)"

func guiWindowBuiltin(args ...Object) Object {
	return newError(guiUnavailableMsg)
}

func guiButtonBuiltin(args ...Object) Object {
	return newError(guiUnavailableMsg)
}

func guiShowBuiltin(args ...Object) Object {
	return newError(guiUnavailableMsg)
}

func guiCloseBuiltin(args ...Object) Object {
	return newError(guiUnavailableMsg)
}

func guiAlertBuiltin(args ...Object) Object {
	return newError(guiUnavailableMsg)
}

func guiTableBuiltin(args ...Object) Object {
	return newError(guiUnavailableMsg)
}

func guiRunBuiltin(args ...Object) Object {
	return newError(guiUnavailableMsg)
}
