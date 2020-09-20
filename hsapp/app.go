package hsapp

import (
	"encoding/json"
	"image/color"
	"io/ioutil"
	"runtime"
	"strconv"

	"github.com/hajimehoshi/ebiten/ebitenutil"

	"github.com/OpenDiablo2/HellSpawner/hscommon"
	"github.com/OpenDiablo2/HellSpawner/hsutil"

	"github.com/OpenDiablo2/HellSpawner/hsconfig"
	"github.com/OpenDiablo2/HellSpawner/hsui"
	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten"
	"golang.org/x/image/font"
)

const (
	bytesToMegabyte          = 1024 * 1024
	windowFrameHeight    int = 30
	windowFrameThickness int = 3
)

// App is an instance of the HellSpawner app.
type App struct {
	Config         hsconfig.AppConfig
	NormalFont     font.Face
	SymbolFont     font.Face
	MonospaceFont  font.Face
	ttNormal       *truetype.Font
	ttMono         *truetype.Font
	ttSymbols      *truetype.Font
	screenWidth    int
	screenHeight   int
	mouseX, mouseY int

	testbox *hsui.VBox
}

func (a *App) GetAppConfig() *hsconfig.AppConfig {
	return &a.Config
}

func (a *App) GetNormalFont() font.Face {
	return a.NormalFont
}

func (a *App) GetSymbolsFont() font.Face {
	return a.SymbolFont
}

func (a *App) GetMonospaceFont() font.Face {
	return a.MonospaceFont
}

// Create creates an instance of the HelSpawner app.
func Create() (*App, error) {
	result := &App{}

	var err error

	var configBytes []byte

	// Load the configuration file
	// TODO: Check for a custom config file
	if configBytes, err = ioutil.ReadFile("config.json"); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(configBytes, &result.Config); err != nil {
		return nil, err
	}

	// Set up the initial window layout and properties
	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowTitle("HellSpawner")
	ebiten.SetWindowResizable(true)
	ebiten.SetVsyncEnabled(true)
	ebiten.SetWindowDecorated(true)

	// Configure the fonts for rendering
	if err = result.configureFonts(); err != nil {
		return nil, err
	}

	// Store off the device scale factor so we can regenerate if we need to
	hsutil.SetDeviceScale(ebiten.DeviceScaleFactor())

	result.testbox = hsui.CreateVBox()
	hbox := hsui.CreateHBox()
	hbox.SetExpandChild(true)
	hbox.AddChild(hsui.CreateButton(result, "Left", func() {}))
	hbox.AddChild(hsui.CreateButton(result, "Center", func() {}))
	hbox.AddChild(hsui.CreateButton(result, "Right", func() {}))

	b1 := hsui.CreateButton(result, "Align Middle", func() { result.testbox.SetAlignment(hscommon.VAlignMiddle) })
	result.testbox.AddChild(hsui.CreateButton(result, "Align Top", func() { result.testbox.SetAlignment(hscommon.VAlignTop) }))
	result.testbox.AddChild(b1)
	result.testbox.AddChild(hsui.CreateButton(result, "Align Bottom", func() { result.testbox.SetAlignment(hscommon.VAlignBottom) }))
	result.testbox.AddChild(hsui.CreateButton(result, "Vis Toggle", func() { b1.SetVisible(!b1.GetVisible()) }))
	result.testbox.AddChild(hsui.CreateButton(result, "Toggle Expand Child", func() { result.testbox.SetExpandChild(!result.testbox.GetExpandChild()) }))
	result.testbox.AddChild(hsui.CreateButton(result, "Child Spacing +", func() { result.testbox.SetChildSpacing(result.testbox.GetChildSpacing() + 1) }))
	result.testbox.AddChild(hsui.CreateButton(result, "Child Spacing -", func() { result.testbox.SetChildSpacing(result.testbox.GetChildSpacing() - 1) }))
	result.testbox.AddChild(hbox)

	return result, nil
}

func (a *App) Run() error {
	return ebiten.RunGame(a)
}

func (a *App) Update(*ebiten.Image) error {
	deviceScale := ebiten.DeviceScaleFactor()
	a.mouseX, a.mouseY = ebiten.CursorPosition()

	// If the device scale has changed, we need to regenerate the fonts
	if deviceScale != hsutil.GetLastDeviceScale() {
		hsutil.SetDeviceScale(deviceScale)
		a.regenerateFonts()
		a.testbox.Invalidate()
	}

	a.testbox.Update()

	return nil
}

func (a *App) Draw(screen *ebiten.Image) {
	frameColor := a.Config.Colors.WindowBackground

	// Fill the window with the frame color
	_ = screen.Fill(color.RGBA{R: frameColor[0], G: frameColor[1], B: frameColor[2], A: frameColor[3]})

	a.testbox.Render(screen, 0, 0, 300, a.screenHeight)

	// Debug print stuff
	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)
	ebitenutil.DebugPrintAt(screen,
		"Alloc:   "+strconv.FormatInt(int64(m.Alloc)/bytesToMegabyte, 10)+"\n"+
			"Pause:   "+strconv.FormatInt(int64(m.PauseTotalNs/bytesToMegabyte), 10)+"\n"+
			"HeapSys: "+strconv.FormatInt(int64(m.HeapSys/bytesToMegabyte), 10)+"\n"+
			"NumGC:   "+strconv.FormatInt(int64(m.NumGC), 10),
		hsutil.ScaleToDevice(550), hsutil.ScaleToDevice(100))
}

func (a *App) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	// Store off the screen size for easy access
	a.screenWidth = outsideWidth
	a.screenHeight = outsideHeight

	// Return the actual resolution, determined by the virtual size and screen device scale
	return hsutil.ScaleToDevice(outsideWidth), hsutil.ScaleToDevice(outsideHeight)
}

// configureFonts loads the fonts the app needs.
func (a *App) configureFonts() error {
	var ttNormalBytes, ttMonoBytes, ttSymbolBytes []byte

	var err error

	// Load font data from the files into a byte array
	if ttNormalBytes, err = ioutil.ReadFile(a.Config.Fonts.Normal.Face); err != nil {
		return err
	}

	if ttSymbolBytes, err = ioutil.ReadFile(a.Config.Fonts.Symbols.Face); err != nil {
		return err
	}

	if ttMonoBytes, err = ioutil.ReadFile(a.Config.Fonts.Monospaced.Face); err != nil {
		return err
	}

	// Parse the TTF font files
	if a.ttNormal, err = truetype.Parse(ttNormalBytes); err != nil {
		return err
	}

	if a.ttSymbols, err = truetype.Parse(ttSymbolBytes); err != nil {
		return err
	}

	if a.ttMono, err = truetype.Parse(ttMonoBytes); err != nil {
		return err
	}

	// Generate the fonts
	a.regenerateFonts()

	return nil
}

// regenerateFonts regenerates the fonts (at startup and when dragging to a different DPI display).
func (a *App) regenerateFonts() {
	deviceScale := ebiten.DeviceScaleFactor()

	a.NormalFont = truetype.NewFace(a.ttNormal, &truetype.Options{
		Size:    float64(a.Config.Fonts.Normal.Size),
		DPI:     96 * deviceScale,
		Hinting: font.HintingNone,
	})

	a.SymbolFont = truetype.NewFace(a.ttSymbols, &truetype.Options{
		Size:    float64(a.Config.Fonts.Symbols.Size),
		DPI:     96 * deviceScale,
		Hinting: font.HintingNone,
	})

	a.MonospaceFont = truetype.NewFace(a.ttMono, &truetype.Options{
		Size:    float64(a.Config.Fonts.Monospaced.Size),
		DPI:     96 * deviceScale,
		Hinting: font.HintingNone,
	})

}
