package main

import (
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/joshuaferrara/go-satellite"
)

func main() {
	// Example TLE for NOAA 19
	//NOAA 19
	// 1 33591U 09005A   25074.18988975  .00000419  00000+0  24768-3 0  9991
	// 2 33591  99.0072 138.3781 0012918 245.4492 114.5334 14.13308947829901

	tleLine1 := "1 33591U 09005A   25074.18988975  .00000419  00000+0  24768-3 0  9991"
	tleLine2 := "2 33591  99.0072 138.3781 0012918 245.4492 114.5334 14.13308947829901"

	// Parse the TLE data into a Satellite object
	sat := satellite.TLEToSat(tleLine1, tleLine2, satellite.GravityWGS72)

	rl.SetConfigFlags(rl.FlagWindowResizable | rl.FlagMsaa4xHint)
	rl.InitWindow(1280, 720, "NOAA 19")
	rl.MaximizeWindow()
	defer rl.CloseWindow()

	earth_model := rl.LoadModel("res/Earth_1_12756.glb")
	satellite_model := rl.LoadModel("res/satellite.glb")

	camera := rl.Camera{}
	camera.Position = rl.NewVector3(-10.0, 8.0, -10.0)
	camera.Target = rl.NewVector3(0.0, 0.0, 0.0)
	camera.Up = rl.NewVector3(0.0, 1.0, 0.0)
	camera.Fovy = 45.0
	camera.Projection = rl.CameraPerspective

	rl.SetTargetFPS(60)

	const scale = 1030.0
	followSatellite := false
	var zoom float32 = 1.0

	for !rl.WindowShouldClose() {
		now := time.Now().UTC()

		position, _ := satellite.Propagate(sat, now.Year(), int(now.Month()), now.Day(), now.Hour(), now.Minute(), now.Second())

		satPos := rl.NewVector3((float32(position.Y / scale)), (float32(position.Z / scale)), (float32(position.X / scale)))

		if rl.IsKeyPressed(rl.KeyF1) {
			followSatellite = !followSatellite
			// Reset the camera position
			camera.Position = rl.NewVector3(-10.0, 8.0, -10.0)
			rl.ShowCursor()
		}

		if followSatellite {
			camera.Position = rl.Vector3{
				X: satPos.X * zoom,
				Y: satPos.Y * zoom,
				Z: satPos.Z * zoom,
			}
			zoom -= rl.GetMouseWheelMoveV().Y / 10
			zoom = rl.Clamp(zoom, 1.1, 10.0)
		} else {
			rl.UpdateCamera(&camera, rl.CameraThirdPerson)
			rl.SetMousePosition(rl.GetScreenWidth() / 2, rl.GetScreenHeight() / 2)
			rl.HideCursor()
		}

		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)

		rl.BeginMode3D(camera)

		rl.DrawGrid(50, 2.0)
		rl.DrawLine3D(rl.NewVector3(0, 0, 0), rl.NewVector3(100, 0, 0), rl.Red)
		rl.DrawLine3D(rl.NewVector3(0, 0, 0), rl.NewVector3(0, 100, 0), rl.Green)
		rl.DrawLine3D(rl.NewVector3(0, 0, 0), rl.NewVector3(0, 0, 100), rl.Blue)

		// Calculate how many seconds have passed in the day
		secondsInDay := float32(now.Hour()*3600 + now.Minute()*60 + now.Second())

		// Rotation angle per second = 360 degrees / 86400 seconds = 0.004167 degrees/second
		rotationAngle := secondsInDay * 360.0 / 86400.0

		rl.DrawModelEx(earth_model, rl.NewVector3(0, 0, 0), rl.NewVector3(0, 1, 0), rotationAngle-5, rl.NewVector3(0.01, 0.01, 0.01), rl.White)

		// Draw the satellite position as a red sphere
		rl.DrawModelEx(satellite_model, satPos, rl.NewVector3(0, 0, 0), 0.0, rl.NewVector3(0.0001, 0.0001, 0.0001), rl.White)
		rl.DrawLine3D(satPos, rl.NewVector3(0, 0, 0), rl.Blue)

		rl.EndMode3D()

		text := now.String()
		rl.DrawText(text, 10, 10, 20, rl.White)

		rl.EndDrawing()
	}

}
