package ui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

var cursorTimer int32
var focus bool

func InputText(label string, x int32, y int32, width int32, height int32, buf *string, maxCharacters uint) {
	textBox := rl.Rectangle{X: float32(x), Y: float32(y), Width: float32(width), Height: float32(height)}

	letterCount := len(*buf)
	mouseOnText := rl.CheckCollisionPointRec(rl.GetMousePosition(), textBox)

	if mouseOnText {
		rl.SetMouseCursor(rl.MouseCursorIBeam)

		if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			focus = true
		}
	} else {
		rl.SetMouseCursor(rl.MouseCursorDefault)

		if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			focus = false
		}
	}

	if focus {
		// Get the character pressed (unicode character)
		key := rl.GetCharPressed()

		// Process characters in the queue
		for key > 0 {
			// Only allow printable characters within the specified size
			if key >= 32 && key <= 125 && letterCount < int(maxCharacters) {
				*buf += string(key) // Append the new character to the buffer
				letterCount++
			}
			key = rl.GetCharPressed() // Check next character in the queue
		}

		// Handle backspace key to delete characters
		if rl.IsKeyPressed(rl.KeyBackspace) && letterCount > 0 {
			letterCount--
			*buf = (*buf)[:letterCount] // Remove last character
		}
	}

	rl.DrawText(label, x, y-20, 20, rl.DarkGray)
	rl.DrawRectangleRec(textBox, rl.LightGray)

	// Highlight the input box if the mouse is over it
	if focus {
		rl.DrawRectangleLinesEx(textBox, 2.0, rl.Yellow)
	}

	cursorTimer++
	if focus && letterCount < int(maxCharacters) {
		if (cursorTimer/20)%2 == 0 {
			rl.DrawText("|", int32(textBox.X)+10+rl.MeasureText(*buf, 40), int32(textBox.Y)+8, 40, rl.DarkBlue)
		}
	}

	// Draw the text entered so far
	rl.DrawText(*buf, int32(textBox.X)+5, int32(textBox.Y)+8, 40, rl.DarkBlue)
}