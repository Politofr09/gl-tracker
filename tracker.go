package main

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/joshuaferrara/go-satellite"

	"gl-tracker/internal/tle"
	"gl-tracker/internal/ui"

	"gopkg.in/ini.v1"
)

func selectSatellite(selectedSatellite string, satellitesList map[string][2]string, scale float64, t time.Time, orbitPath []rl.Vector3, orbitPoints int) (satellite.Satellite, error) {
	if _, exists := satellitesList[selectedSatellite]; !exists {
		return satellite.Satellite{}, errors.New("Can't find satellite " + selectedSatellite)
	}

	tleLine1 := satellitesList[selectedSatellite][0]
	tleLine2 := satellitesList[selectedSatellite][1]

	// Parse the TLE data into a Satellite object
	sat := satellite.TLEToSat(tleLine1, tleLine2, satellite.GravityWGS84)
	computeOrbitPath(sat, t, scale, orbitPath[:], orbitPoints)
	rl.SetWindowTitle("Tracking " + selectedSatellite)
	return sat, nil
}

func getSatellitePosition(sat satellite.Satellite, time time.Time, scale float64) rl.Vector3 {
	position, _ := satellite.Propagate(sat, time.Year(), int(time.Month()), time.Day(), time.Hour(), time.Minute(), time.Second())

	position = satellite.ECIToECEF(position, satellite.GSTimeFromDate(time.Year(), int(time.Month()), time.Day(), time.Hour(), time.Minute(), time.Second()))

	return rl.NewVector3((float32(position.Y / scale)), (float32(position.Z / scale)), (float32(position.X / scale)))
}

func computeOrbitPath(sat satellite.Satellite, t time.Time, scale float64, orbitPath []rl.Vector3, orbitPoints int) {

	if orbitPoints <= 10 {
		t = t.Add(-time.Duration(orbitPoints) * time.Minute / 2) // Remove the half
	} else {
		t = t.Add(time.Minute * -10) // Remove 10 minutes
	}

	for i := 0; i < orbitPoints; i++ {
		futureTime := t.Add(time.Second * time.Duration(i*60)) // Every 60s
		orbitPath[i] = getSatellitePosition(sat, futureTime, scale)
	}
}

func main() {
	// Load config
	cfg, err := ini.Load("res/config.ini")
	if err != nil {
		fmt.Printf("Error reading config file: %v\n", err)
		return
	}

	const scale = 1050.0
	followSatellite := false
	var zoom float32 = 1.0

	err = tle.FetchTLEs()
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
	rl.SetTraceLogLevel(rl.LogNone)
	rl.InitWindow(1280, 720, "No satellite selected yet")
	rl.MaximizeWindow()
	defer rl.CloseWindow()

	selectedSatellite := cfg.Section("Tracker").Key("satellite").String()
	fmt.Println("[Config] Selected satellite:", selectedSatellite)

	// Parse lat/long from
	parts := strings.Split(cfg.Section("Tracker").Key("ground_station").String(), ", ")
	if len(parts) < 2 {
		fmt.Printf("Invalid ground station format")
		return
	}

	groundLat, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		fmt.Printf("Invalid latitude: %v", err)
		return
	}

	groundLon, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		fmt.Printf("Invalid longitude: %v", err)
		return
	}
	fmt.Println("[Config] Ground station:", groundLon, groundLat)

	now := time.Now().UTC()

	inputText := ""
	orbitPoints := cfg.Section("Tracker").Key("orbit_duration").MustInt(100)
	orbitStyle := cfg.Section("Tracker").Key("orbit_style").String()
	fmt.Println("[Config] Orbit style:", orbitStyle)
	var drawOrbitAsPoint bool
	if orbitStyle == "Points" {
		drawOrbitAsPoint = true
	} else if orbitStyle == "Lines" {
		drawOrbitAsPoint = false
	} else {
		fmt.Printf("Invalid orbit style: %v", orbitStyle)
		fmt.Println("Expected 'Points' or 'Lines'")
		return
	}

	fmt.Println("[Config] Orbit duration in minutes:", orbitPoints)

	var orbitPath = make([]rl.Vector3, orbitPoints)
	sat, err := selectSatellite(selectedSatellite, satellites, scale, now, orbitPath[:], orbitPoints)
	if err != nil {
		fmt.Printf("Problem selecting satellite: %q\n", err)
	}

	// Load shaders
	crtShader := rl.LoadShader("res/crt.vs", "res/crt.fs")
	timeLoc := rl.GetShaderLocation(crtShader, "time")
	resLoc := rl.GetShaderLocation(crtShader, "resolution")
	defer rl.UnloadShader(crtShader)

	earthModel := rl.LoadModel("res/Earth_1_12756.glb")
	defer rl.UnloadModel(earthModel)
	satelliteModel := rl.LoadModel("res/satellite.glb")
	defer rl.UnloadModel(satelliteModel)

	bilboardTexture := rl.LoadTexture("res/location-bilboard-map.png")
	defer rl.UnloadTexture(bilboardTexture)

	ui.HackerFont = rl.LoadFontEx("res/ShareTechMono-Regular.ttf", 40, nil)
	rl.SetTextureFilter(ui.HackerFont.Texture, rl.FilterBilinear)
	defer rl.UnloadFont(ui.HackerFont)

	renderTarget := rl.LoadRenderTexture(int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight()))
	camera := rl.Camera{}
	camera.Position = rl.NewVector3(-10.0, 8.0, -10.0)
	camera.Target = rl.NewVector3(0.0, 0.0, 0.0)
	camera.Up = rl.NewVector3(0.0, 1.0, 0.0)
	camera.Fovy = 45.0
	camera.Projection = rl.CameraPerspective

	rl.SetTargetFPS(60)

	// now := time.Now().UTC()
	for !rl.WindowShouldClose() {
		// now = now.Add(time.Second * 10)
		now = time.Now().UTC()

		computeOrbitPath(sat, now, scale, orbitPath[:], orbitPoints)

		// Recreate the render texture if the window is resized
		if rl.IsWindowResized() {
			rl.UnloadRenderTexture(renderTarget)
			renderTarget = rl.LoadRenderTexture(int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight()))
		}

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

		rl.DrawGrid(50, 10)

		rl.DrawModelEx(earthModel, rl.NewVector3(0, 0, 0), rl.NewVector3(0, 1, 0), 180, rl.NewVector3(0.01, 0.01, 0.01), rl.White)

		// Draw orbit
		if drawOrbitAsPoint {
			for i := 0; i < orbitPoints; i++ {
				rl.DrawPoint3D(orbitPath[i], ui.GREEN_HACKER_COLOR)
			}
		} else {
			for i := 0; i < orbitPoints-1; i++ {
				rl.DrawLine3D(orbitPath[i], orbitPath[i+1], ui.GREEN_HACKER_COLOR)
				// rl.DrawPoint3D(orbitPath[i], ui.GREEN_HACKER_COLOR)
			}
		}
		rl.DrawModelEx(satelliteModel, satPos, rl.NewVector3(0, 0, 0), 0.0, rl.NewVector3(0.0001, 0.0001, 0.0001), rl.White)
		rl.DrawLine3D(satPos, rl.NewVector3(0, 0, 0), ui.GREEN_HACKER_COLOR)

		// Draw ground station
		groundStationECI := satellite.LLAToECI(satellite.LatLong{Latitude: groundLat * satellite.DEG2RAD, Longitude: groundLon * satellite.DEG2RAD}, 0.0, satellite.JDay(now.Year(), int(now.Month()), now.Day(), now.Hour(), now.Minute(), now.Second()))
		groundStation := satellite.ECIToECEF(groundStationECI, satellite.GSTimeFromDate(now.Year(), int(now.Month()), now.Day(), now.Hour(), now.Minute(), now.Second()))

		rotationMatrix := rl.MatrixRotateY(rl.Deg2rad * 90)
		groundStationRotated := rl.Vector3Transform(rl.NewVector3(float32(groundStation.X), float32(groundStation.Z), -float32(groundStation.Y)), rotationMatrix)
		// rl.DrawSphere(rl.NewVector3(groundStationRotated.X/1300, groundStationRotated.Y/1300, groundStationRotated.Z/1300), 0.1, rl.White)
		rl.DrawBillboard(camera, bilboardTexture, rl.NewVector3(groundStationRotated.X/1260, groundStationRotated.Y/1260+0.1, groundStationRotated.Z/1260), 0.5, ui.GREEN_HACKER_COLOR)

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
		dateText := now.String()
		rl.DrawTextEx(ui.HackerFont, dateText, rl.NewVector2(10, 10), 20, 1.0, ui.GREEN_HACKER_COLOR)

		helpText := "F1: Follow satellite"
		if followSatellite {
			helpText = "F1: Globe view"
		}
		rl.DrawTextEx(ui.HackerFont, helpText, rl.NewVector2(float32(rl.GetScreenWidth())-rl.MeasureTextEx(ui.HackerFont, helpText, 20, 1.0).X-10, 10), 20, 1.0, ui.GREEN_HACKER_COLOR)

		ui.InputText("Select a satellite", 10, 80, 400, 50, &inputText, 20)
		inputText = strings.ToUpper(inputText)
		if inputText != selectedSatellite && inputText != "" {
			tempSat, err := selectSatellite(inputText, satellites, scale, now, orbitPath[:], orbitPoints)
			if err == nil {
				selectedSatellite = inputText
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
				rl.DrawTextEx(ui.HackerFont, slice, rl.NewVector2(10, 140+float32(i)*20), 20, 1.0, rl.White)
			}
		}

		// Draw text for the selected satellite
		satNamePos := rl.GetWorldToScreen(satPos, camera)
		rl.DrawTextEx(ui.HackerFont, selectedSatellite, rl.NewVector2(satNamePos.X-rl.MeasureTextEx(ui.HackerFont, selectedSatellite, 20, 1.0).X/2, satNamePos.Y), 20, 1.0, ui.GREEN_HACKER_COLOR)

		rl.EndDrawing()
	}

}
