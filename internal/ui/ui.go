package ui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

func InputText(label string, x int32, y int32, width int32, height int32, buf *string, maxCharacters uint) {
	textBox := rl.Rectangle{X: float32(x), Y: float32(y), Width: float32(width), Height: float32(height)}

	letterCount := len(*buf)
	mouseOnText := false

	if rl.CheckCollisionPointRec(rl.GetMousePosition(), textBox) {
		mouseOnText = true
	} else {
		mouseOnText = false
	}

	if mouseOnText {
		// Set the window's cursor to the I-Beam for text input
		rl.SetMouseCursor(rl.MouseCursorIBeam)

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
	} else {
		rl.SetMouseCursor(rl.MouseCursorDefault)
	}

	rl.DrawText(label, x, y-20, 20, rl.DarkGray)
	rl.DrawRectangleRec(textBox, rl.LightGray)

	// Highlight the input box if the mouse is over it
	if mouseOnText {
		rl.DrawRectangleLines(int32(textBox.X), int32(textBox.Y), int32(textBox.Width), int32(textBox.Height), rl.Red)
	} else {
		rl.DrawRectangleLines(int32(textBox.X), int32(textBox.Y), int32(textBox.Width), int32(textBox.Height), rl.DarkGray)
	}

	// Draw the text entered so far
	rl.DrawText(*buf, int32(textBox.X)+5, int32(textBox.Y)+8, 40, rl.Maroon)


}