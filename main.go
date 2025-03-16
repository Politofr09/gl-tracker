package main

import (
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/joshuaferrara/go-satellite"
)

func getSatellitePosition(sat satellite.Satellite, time time.Time, scale float64) rl.Vector3 {
	position, _ := satellite.Propagate(sat, time.Year(), int(time.Month()), time.Day(), time.Hour(), time.Minute(), time.Second())

	return rl.NewVector3((float32(position.Y / scale)), (float32(position.Z / scale)), (float32(position.X / scale)))
}

func computeOrbitPath(sat satellite.Satellite, scale float64, orbitPath []rl.Vector3, orbitPoints int) {
	now := time.Now().UTC().Add(time.Minute * -10) // Remove 10 minutes
	for i := 0; i < orbitPoints; i++ {
		futureTime := now.Add(time.Second * time.Duration(i*60)) // Every 60s
		position, _ := satellite.Propagate(sat, futureTime.Year(), int(futureTime.Month()), futureTime.Day(), futureTime.Hour(), futureTime.Minute(), futureTime.Second())

		// Convert to scaled world space
		orbitPath[i] = rl.NewVector3(
			float32(position.Y/scale),
			float32(position.Z/scale),
			float32(position.X/scale),
		)
	}
}

func main() {
	const scale = 1050.0
	followSatellite := false
	var zoom float32 = 1.0

	// Example TLE for NOAA 19
	//NOAA 19
	// 1 33591U 09005A   25074.18988975  .00000419  00000+0  24768-3 0  9991
	// 2 33591  99.0072 138.3781 0012918 245.4492 114.5334 14.13308947829901

	// NOAA 18
	// 1 28654U 05018A   25075.18855678  .00000469  00000+0  27229-3 0  9999
	// 2 28654  98.8462 155.6171 0014190  16.3868 343.7760 14.13542600 21680

	// GOES 14
	// 1 35491U 09033A   25074.65733697 -.00000069  00000+0  00000+0 0  9995
	// 2 35491   0.4919  89.3626 0000671 182.2428  30.4186  1.00269709  1997

	tleLine1 := "1 28654U 05018A   25075.18855678  .00000469  00000+0  27229-3 0  9999"
	tleLine2 := "2 28654  98.8462 155.6171 0014190  16.3868 343.7760 14.13542600 21680"

	// Parse the TLE data into a Satellite object
	sat := satellite.TLEToSat(tleLine1, tleLine2, satellite.GravityWGS72)

	// Compute the orbit
	const orbitPoints = 110
	var orbitPath [orbitPoints]rl.Vector3

	computeOrbitPath(sat, scale, orbitPath[:], orbitPoints)

	rl.SetConfigFlags(rl.FlagWindowResizable | rl.FlagMsaa4xHint)
	rl.InitWindow(1280, 720, "NOAA 18")
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

	for !rl.WindowShouldClose() {
		now := time.Now().UTC()

		satPos := getSatellitePosition(sat, now, scale)

		if rl.IsKeyPressed(rl.KeyF1) {
			followSatellite = !followSatellite
			// Reset the camera position
			camera.Position = rl.NewVector3(-10.0, 8.0, -10.0)
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
			rl.UpdateCamera(&camera, rl.CameraOrbital)
		}

		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)

		rl.BeginMode3D(camera)

		rl.DrawGrid(50, 6378/1000.0)

		// Calculate how many seconds have passed in the day
		secondsInDay := float32(now.Hour()*3600 + now.Minute()*60 + now.Second())

		// Rotation angle per second = 360 degrees / 86400 seconds = 0.004167 degrees/second
		rotationAngle := secondsInDay * 360.0 / 86400.0

		rl.DrawModelEx(earth_model, rl.NewVector3(0, 0, 0), rl.NewVector3(0, 1, 0), rotationAngle-5, rl.NewVector3(0.01, 0.01, 0.01), rl.White)

		// Draw orbit
		for i := 0; i < orbitPoints-1; i++ {
			rl.DrawLine3D(orbitPath[i], orbitPath[i+1], rl.SkyBlue)
		}

		// Draw the satellite position as a red sphere
		rl.DrawModelEx(satellite_model, satPos, rl.NewVector3(0, 0, 0), 0.0, rl.NewVector3(0.0001, 0.0001, 0.0001), rl.White)
		rl.DrawLine3D(satPos, rl.NewVector3(0, 0, 0), rl.Blue)

		rl.EndMode3D()

		text := now.String()
		rl.DrawText(text, 10, 10, 20, rl.White)

		rl.EndDrawing()
	}

}
