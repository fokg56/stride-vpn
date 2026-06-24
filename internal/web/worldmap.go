package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type MapCountry struct {
	Name  string    `json:"name"`
	Path  string    `json:"path"`
	Label [2]string `json:"label"`
	X     float64   `json:"x"`
	Y     float64   `json:"y"`
}

type MapData struct {
	ViewBox   [4]float64  `json:"viewBox"`
	Countries []MapCountry `json:"countries"`
}

type coord struct{ x, y float64 }

func svgPath(coords ...coord) string {
	if len(coords) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("M%.1f,%.1f", coords[0].x, coords[0].y))
	for _, c := range coords[1:] {
		b.WriteString(fmt.Sprintf("L%.1f,%.1f", c.x, c.y))
	}
	b.WriteString("Z")
	return b.String()
}

func lonlat(lon, lat float64) coord {
	return coord{
		x: (lon + 180) / 360 * 800,
		y: (90 - lat) / 180 * 400,
	}
}

func worldMapData() MapData {
	// North America
	na := svgPath(
		lonlat(-168, 54), lonlat(-160, 58), lonlat(-152, 62), lonlat(-142, 66),
		lonlat(-132, 68), lonlat(-122, 70), lonlat(-110, 70), lonlat(-100, 72),
		lonlat(-90, 70), lonlat(-82, 72), lonlat(-74, 68), lonlat(-66, 62),
		lonlat(-60, 56), lonlat(-56, 52), lonlat(-54, 48), lonlat(-56, 46),
		lonlat(-60, 44), lonlat(-65, 42), lonlat(-70, 40), lonlat(-75, 36),
		lonlat(-78, 34), lonlat(-80, 32), lonlat(-80, 28), lonlat(-80, 26),
		lonlat(-82, 24), lonlat(-84, 22), lonlat(-88, 20), lonlat(-90, 18),
		lonlat(-88, 16), lonlat(-85, 14), lonlat(-82, 12), lonlat(-80, 10),
		lonlat(-78, 8), lonlat(-80, 8), lonlat(-83, 10), lonlat(-86, 12),
		lonlat(-90, 14), lonlat(-94, 16), lonlat(-98, 18), lonlat(-102, 20),
		lonlat(-106, 23), lonlat(-110, 25), lonlat(-112, 28), lonlat(-115, 31),
		lonlat(-118, 34), lonlat(-120, 37), lonlat(-122, 40), lonlat(-124, 43),
		lonlat(-125, 46), lonlat(-126, 49), lonlat(-128, 52), lonlat(-132, 55),
		lonlat(-136, 58), lonlat(-142, 60), lonlat(-148, 60), lonlat(-154, 58),
		lonlat(-160, 56), lonlat(-168, 54),
	)

	// South America
	sa := svgPath(
		lonlat(-80, 8), lonlat(-76, 11), lonlat(-72, 13), lonlat(-68, 12),
		lonlat(-64, 11), lonlat(-60, 7), lonlat(-55, 4), lonlat(-52, 2),
		lonlat(-48, 0), lonlat(-44, -3), lonlat(-40, -7), lonlat(-37, -10),
		lonlat(-36, -13), lonlat(-35, -15), lonlat(-37, -18), lonlat(-40, -21),
		lonlat(-43, -24), lonlat(-46, -27), lonlat(-50, -30), lonlat(-54, -33),
		lonlat(-56, -35), lonlat(-58, -37), lonlat(-62, -40), lonlat(-65, -43),
		lonlat(-67, -46), lonlat(-68, -49), lonlat(-70, -52), lonlat(-72, -50),
		lonlat(-74, -46), lonlat(-75, -42), lonlat(-73, -38), lonlat(-72, -34),
		lonlat(-71, -30), lonlat(-73, -25), lonlat(-76, -20), lonlat(-79, -15),
		lonlat(-80, -10), lonlat(-80, -5), lonlat(-79, 0), lonlat(-78, 4),
		lonlat(-80, 8),
	)

	// Europe (simplified outline from Iberia to Urals, including Scandinavia)
	eu := svgPath(
		lonlat(-10, 52), lonlat(-8, 54), lonlat(-5, 57), lonlat(-3, 59),
		lonlat(0, 61), lonlat(3, 63), lonlat(6, 65), lonlat(9, 67),
		lonlat(12, 66), lonlat(15, 64), lonlat(18, 62), lonlat(21, 60),
		lonlat(24, 58), lonlat(27, 56), lonlat(30, 54), lonlat(33, 52),
		lonlat(36, 50), lonlat(40, 48), lonlat(44, 46), lonlat(48, 44),
		lonlat(50, 42), lonlat(48, 40), lonlat(44, 38), lonlat(40, 36),
		lonlat(36, 34), lonlat(32, 32), lonlat(28, 30), lonlat(24, 28),
		lonlat(20, 26), lonlat(16, 24), lonlat(12, 22), lonlat(8, 20),
		lonlat(4, 18), lonlat(0, 16), lonlat(-4, 14), lonlat(-8, 16),
		lonlat(-10, 18), lonlat(-10, 22), lonlat(-8, 26), lonlat(-6, 30),
		lonlat(-4, 34), lonlat(-4, 38), lonlat(-2, 42), lonlat(0, 44),
		lonlat(2, 46), lonlat(0, 48), lonlat(-4, 48), lonlat(-8, 48),
		lonlat(-10, 50), lonlat(-10, 52),
	)

	// Africa
	af := svgPath(
		lonlat(-17, 32), lonlat(-13, 35), lonlat(-8, 36), lonlat(-4, 35),
		lonlat(0, 33), lonlat(5, 32), lonlat(10, 32), lonlat(15, 32),
		lonlat(20, 32), lonlat(25, 32), lonlat(30, 30), lonlat(35, 28),
		lonlat(38, 24), lonlat(42, 20), lonlat(46, 15), lonlat(50, 10),
		lonlat(48, 5), lonlat(44, 0), lonlat(40, -5), lonlat(36, -9),
		lonlat(32, -12), lonlat(28, -16), lonlat(24, -20), lonlat(20, -24),
		lonlat(18, -28), lonlat(16, -32), lonlat(15, -35), lonlat(12, -34),
		lonlat(8, -30), lonlat(4, -25), lonlat(0, -20), lonlat(-4, -15),
		lonlat(-8, -10), lonlat(-12, -5), lonlat(-16, 0), lonlat(-16, 5),
		lonlat(-16, 10), lonlat(-16, 15), lonlat(-15, 20), lonlat(-14, 25),
		lonlat(-13, 28), lonlat(-12, 30), lonlat(-17, 32),
	)

	// Asia (from Urals east to Pacific, including India and Arabia)
	as := svgPath(
		lonlat(30, 50), lonlat(35, 55), lonlat(40, 60), lonlat(45, 65),
		lonlat(55, 68), lonlat(65, 70), lonlat(75, 72), lonlat(85, 72),
		lonlat(95, 72), lonlat(105, 72), lonlat(115, 70), lonlat(125, 68),
		lonlat(135, 66), lonlat(145, 64), lonlat(155, 62), lonlat(165, 60),
		lonlat(175, 58), lonlat(180, 56), lonlat(175, 54), lonlat(165, 52),
		lonlat(155, 50), lonlat(148, 48), lonlat(142, 46), lonlat(136, 44),
		lonlat(132, 42), lonlat(128, 40), lonlat(124, 38), lonlat(120, 36),
		lonlat(116, 34), lonlat(112, 32), lonlat(108, 30), lonlat(104, 28),
		lonlat(100, 26), lonlat(96, 24), lonlat(92, 22), lonlat(88, 20),
		lonlat(84, 18), lonlat(80, 16), lonlat(76, 14), lonlat(72, 12),
		lonlat(68, 10), lonlat(64, 8), lonlat(60, 6), lonlat(56, 4),
		lonlat(52, 2), lonlat(48, 0), lonlat(44, -2), lonlat(40, 0),
		lonlat(36, 2), lonlat(32, 4), lonlat(28, 6), lonlat(24, 8),
		lonlat(20, 10), lonlat(16, 12), lonlat(12, 14), lonlat(8, 16),
		lonlat(10, 18), lonlat(14, 20), lonlat(18, 22), lonlat(22, 24),
		lonlat(26, 26), lonlat(28, 28), lonlat(28, 30), lonlat(26, 32),
		lonlat(26, 34), lonlat(28, 36), lonlat(30, 38), lonlat(30, 40),
		lonlat(30, 42), lonlat(30, 44), lonlat(30, 46), lonlat(30, 48),
		lonlat(30, 50),
	)

	// Australia/Oceania
	oc := svgPath(
		lonlat(114, -22), lonlat(118, -20), lonlat(124, -18), lonlat(130, -16),
		lonlat(136, -14), lonlat(142, -16), lonlat(146, -20), lonlat(148, -24),
		lonlat(150, -28), lonlat(152, -32), lonlat(154, -36), lonlat(156, -40),
		lonlat(158, -44), lonlat(160, -48), lonlat(162, -50), lonlat(160, -48),
		lonlat(156, -44), lonlat(152, -40), lonlat(148, -36), lonlat(144, -32),
		lonlat(140, -28), lonlat(136, -24), lonlat(132, -22), lonlat(128, -20),
		lonlat(124, -20), lonlat(120, -22), lonlat(116, -24), lonlat(114, -22),
	)

	return MapData{
		ViewBox: [4]float64{0, 0, 800, 400},
		Countries: []MapCountry{
			{Name: "NA", Path: na, Label: [2]string{"NA", "North America"}, X: 150, Y: 100},
			{Name: "SA", Path: sa, Label: [2]string{"SA", "South America"}, X: 245, Y: 240},
			{Name: "EU", Path: eu, Label: [2]string{"EU", "Europe"}, X: 360, Y: 85},
			{Name: "AF", Path: af, Label: [2]string{"AF", "Africa"}, X: 365, Y: 155},
			{Name: "AS", Path: as, Label: [2]string{"AS", "Asia"}, X: 540, Y: 65},
			{Name: "OC", Path: oc, Label: [2]string{"OC", "Oceania"}, X: 630, Y: 230},
		},
	}
}

func handleMapData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	json.NewEncoder(w).Encode(worldMapData())
}
