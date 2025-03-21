package main

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/joshuaferrara/go-satellite"

	"gl-tracker/internal/tle"
	"gl-tracker/internal/ui"
)

func selectSatellite(selectedSatellite string, satellitesList map[string][2]string, scale float64, orbitPath []rl.Vector3, orbitPoints int) (satellite.Satellite, error) {
	// Check if selectedSatellite is valid
	actualName := strings.ToUpper(selectedSatellite)

	if _, exists := satellitesList[actualName]; !exists {
		return satellite.Satellite{}, errors.New("Can't find satellite " + actualName)
	}

	tleLine1 := satellitesList[actualName][0]
	tleLine2 := satellitesList[actualName][1]

	// Parse the TLE data into a Satellite object
	sat := satellite.TLEToSat(tleLine1, tleLine2, satellite.GravityWGS84)
	computeOrbitPath(sat, scale, orbitPath[:], orbitPoints)
	rl.SetWindowTitle("Tracking " + actualName)
	return sat, nil
}

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

	err := tle.FetchTLEs()
	if err != nil {
		fmt.Println("Error fetching TLE data: ", err)
		return
	}

	satellites, err := tle.LoadTLEs()
	if err != nil {
		fmt.Println("Error loading TLEs:", err)
		return
	}

	rl.SetConfigFlags(rl.FlagWindowResizable | rl.FlagMsaa4xHint)
	rl.InitWindow(1280, 720, "No satellite selected yet")
	rl.MaximizeWindow()
	defer rl.CloseWindow()

	selectedSatellite := "NOAA 19"
	inputText := ""
	const orbitPoints = 110
	var orbitPath [orbitPoints]rl.Vector3
	sat, _ := selectSatellite(selectedSatellite, satellites, scale, orbitPath[:], orbitPoints)


	// Load shaders
	crtShader := rl.LoadShader("res/crt.vs", "res/crt.fs")
	timeLoc := rl.GetShaderLocation(crtShader, "time")
	resLoc := rl.GetShaderLocation(crtShader, "resolution")
	defer rl.UnloadShader(crtShader)

	earth_model := rl.LoadModel("res/Earth_1_12756.glb")
	defer rl.UnloadModel(earth_model)
	satellite_model := rl.LoadModel("res/satellite.glb")
	defer rl.UnloadModel(satellite_model)

	renderTarget := rl.LoadRenderTexture(int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight()))
	camera := rl.Camera{}
	camera.Position = rl.NewVector3(-10.0, 8.0, -10.0)
	camera.Target = rl.NewVector3(0.0, 0.0, 0.0)
	camera.Up = rl.NewVector3(0.0, 1.0, 0.0)
	camera.Fovy = 45.0
	camera.Projection = rl.CameraPerspective

	rl.SetTargetFPS(60)

	for !rl.WindowShouldClose() {
		now := time.Now().UTC()
		rl.SetShaderValue(crtShader, timeLoc, []float32{float32(rl.GetTime())}, rl.ShaderUniformFloat)
		rl.SetShaderValue(crtShader, resLoc, []float32{float32(rl.GetScreenWidth()), float32(rl.GetScreenHeight())}, rl.ShaderUniformVec2)

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

		rl.BeginTextureMode(renderTarget)

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

		rl.DrawModelEx(satellite_model, satPos, rl.NewVector3(0, 0, 0), 0.0, rl.NewVector3(0.0001, 0.0001, 0.0001), rl.White)
		rl.DrawLine3D(satPos, rl.NewVector3(0, 0, 0), rl.Blue)

		rl.EndMode3D()

		rl.EndTextureMode()
		
		rl.BeginDrawing()
		rl.BeginShaderMode(crtShader)

		rl.DrawTextureRec(
			renderTarget.Texture, 
			rl.NewRectangle(0, 0, float32(rl.GetScreenWidth()), -float32(rl.GetScreenHeight())), // Flip vertically
			rl.NewVector2(0, 0), // Adjust position accordingly
			rl.White,
		)
		rl.EndShaderMode()

		
		// Render the UI
		date_text := now.String()
		rl.DrawText(date_text, 10, 10, 20, rl.Yellow)

		help_text := "F1: Follow satellite"
		rl.DrawText(help_text, int32(rl.GetScreenWidth())-rl.MeasureText(help_text, 20)-10, 10, 20, rl.Yellow)

		ui.InputText("Select a satellite", 10, 80, 400, 50, &inputText, 20)
		if inputText != selectedSatellite {
			selectedSatellite = inputText
			tempSat, err := selectSatellite(selectedSatellite, satellites, scale, orbitPath[:], orbitPoints)
			if err == nil {
				sat = tempSat
			} // Don't need to handle the error here

		}

		if inputText != "" {
			var slices []string
			for possible_satellite := range satellites {
				if strings.Contains(possible_satellite, strings.ToUpper(inputText)) {
					slices = append(slices, possible_satellite)
				}
			}

			// In go maps do not guarantee order, so we need to sort the slices alphabetically
			sort.Strings(slices)

			for i, slice := range slices {
				rl.DrawText(slice, 10, 140+int32(i)*20, 20, rl.White)
			}
		}

		rl.EndDrawing()
	}

}
