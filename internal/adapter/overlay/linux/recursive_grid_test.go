//go:build linux && cgo

package linux

import (
	"image"
	"math"
	"testing"

	"github.com/y3owk1n/neru/internal/adapter/overlay/render/badge"
	recursivegridcomponent "github.com/y3owk1n/neru/internal/adapter/overlay/render/recursivegrid"
	"github.com/y3owk1n/neru/internal/config"
	"github.com/y3owk1n/neru/internal/domain"
)

func TestLinuxOverlay_DrawRecursiveGrid_SingleLineByDefault(t *testing.T) {
	t.Parallel()

	mgr, surface := recordingManager()
	bounds := image.Rect(0, 0, 100, 100)
	dims := domain.GridDimensions{Cols: 2, Rows: 2}

	cfg := config.DefaultConfig().RecursiveGrid
	style := recursivegridcomponent.BuildStyle(cfg, fixedTheme(false))

	mgr.x11.DrawRecursiveGridWithSubKeyPreview(
		bounds, 0, "abcd", dims, "", domain.GridDimensions{},
		style, recursivegridcomponent.VirtualPointerState{}, false, 0,
	)

	// 2x2 grid should draw exactly 4 rectangles when secondary line is disabled.
	if len(surface.rects) != 4 {
		t.Fatalf("surface.rects count = %d, want 4", len(surface.rects))
	}

	for i, r := range surface.rects {
		if r.border != style.LineColorARGB() {
			t.Errorf("rect[%d].border = %#08x, want %#08x", i, r.border, style.LineColorARGB())
		}
		if r.lineWidth != style.LineWidthF() {
			t.Errorf("rect[%d].lineWidth = %v, want %v", i, r.lineWidth, style.LineWidthF())
		}
	}
}

func TestLinuxOverlay_DrawRecursiveGrid_SecondaryLine(t *testing.T) {
	t.Parallel()

	mgr, surface := recordingManager()
	bounds := image.Rect(0, 0, 200, 200)
	dims := domain.GridDimensions{Cols: 2, Rows: 2}

	cfg := config.DefaultConfig().RecursiveGrid
	cfg.UI.LineWidth = 2
	cfg.UI.LineColor = config.Color{Light: "#ff0000", Dark: "#ff0000"}
	cfg.UI.SecondaryLineWidth = 4
	cfg.UI.SecondaryLineColor = config.Color{Light: "#00ff00", Dark: "#00ff00"}

	style := recursivegridcomponent.BuildStyle(cfg, fixedTheme(false))
	if !style.HasSecondaryLine() {
		t.Fatal("style.HasSecondaryLine() = false, want true")
	}

	mgr.x11.DrawRecursiveGridWithSubKeyPreview(
		bounds, 0, "abcd", dims, "", domain.GridDimensions{},
		style, recursivegridcomponent.VirtualPointerState{}, false, 0,
	)

	// 2x2 grid with dual lines should draw 2 rectangles per cell = 8 rectangles total.
	if len(surface.rects) != 8 {
		t.Fatalf("surface.rects count = %d, want 8", len(surface.rects))
	}

	expectedPrimaryColor := badge.ParseHexARGB("#ff0000")
	expectedSecondaryColor := badge.ParseHexARGB("#00ff00")
	expectedOffset := int(math.Round((2.0 + 4.0) / 2.0))

	for i := 0; i < 4; i++ {
		primary := surface.rects[i*2]
		secondary := surface.rects[i*2+1]

		if primary.border != expectedPrimaryColor {
			t.Errorf("cell[%d] primary border = %#08x, want %#08x", i, primary.border, expectedPrimaryColor)
		}
		if primary.lineWidth != 2.0 {
			t.Errorf("cell[%d] primary lineWidth = %v, want 2.0", i, primary.lineWidth)
		}

		if secondary.border != expectedSecondaryColor {
			t.Errorf("cell[%d] secondary border = %#08x, want %#08x", i, secondary.border, expectedSecondaryColor)
		}
		if secondary.lineWidth != 4.0 {
			t.Errorf("cell[%d] secondary lineWidth = %v, want 4.0", i, secondary.lineWidth)
		}
		if secondary.fill != 0 {
			t.Errorf("cell[%d] secondary fill = %#08x, want 0 (transparent)", i, secondary.fill)
		}

		expectedSecondaryBounds := primary.bounds.Inset(expectedOffset)
		if secondary.bounds != expectedSecondaryBounds {
			t.Errorf("cell[%d] secondary bounds = %v, want %v (inset by %d)",
				i, secondary.bounds, expectedSecondaryBounds, expectedOffset)
		}
	}
}

func TestLinuxOverlay_DrawRecursiveGrid_SecondaryLineInheritsPrimaryWidthWhenZero(t *testing.T) {
	t.Parallel()

	mgr, surface := recordingManager()
	bounds := image.Rect(0, 0, 200, 200)
	dims := domain.GridDimensions{Cols: 2, Rows: 2}

	cfg := config.DefaultConfig().RecursiveGrid
	cfg.UI.LineWidth = 3
	cfg.UI.LineColor = config.Color{Light: "#000000", Dark: "#000000"}
	cfg.UI.SecondaryLineWidth = 0 // 0 means inherit primary LineWidth
	cfg.UI.SecondaryLineColor = config.Color{Light: "#ffffff", Dark: "#ffffff"}

	style := recursivegridcomponent.BuildStyle(cfg, fixedTheme(false))
	if !style.HasSecondaryLine() {
		t.Fatal("style.HasSecondaryLine() = false, want true")
	}

	mgr.x11.DrawRecursiveGridWithSubKeyPreview(
		bounds, 0, "abcd", dims, "", domain.GridDimensions{},
		style, recursivegridcomponent.VirtualPointerState{}, false, 0,
	)

	if len(surface.rects) != 8 {
		t.Fatalf("surface.rects count = %d, want 8", len(surface.rects))
	}

	expectedOffset := int(math.Round((3.0 + 3.0) / 2.0)) // 3px inset

	for i := 0; i < 4; i++ {
		primary := surface.rects[i*2]
		secondary := surface.rects[i*2+1]

		if secondary.lineWidth != 3.0 {
			t.Errorf("cell[%d] secondary lineWidth = %v, want 3.0 (inherited)", i, secondary.lineWidth)
		}
		expectedSecondaryBounds := primary.bounds.Inset(expectedOffset)
		if secondary.bounds != expectedSecondaryBounds {
			t.Errorf("cell[%d] secondary bounds = %v, want %v", i, secondary.bounds, expectedSecondaryBounds)
		}
	}
}
