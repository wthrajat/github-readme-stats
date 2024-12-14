package common

import (
	"strings"
	"testing"
)

func TestKFormatter(t *testing.T) {
	cases := map[float64]string{
		0: "0", 500: "500", 999: "999",
		1000: "1k", 1500: "1.5k", 6600: "6.6k", 10000: "10k",
		-1200: "-1.2k",
	}
	for in, want := range cases {
		if got := KFormatter(in); got != want {
			t.Errorf("KFormatter(%v)=%q want %q", in, got, want)
		}
	}
}

func TestIsValidHexColor(t *testing.T) {
	valid := []string{"fff", "2f80ed", "ffffff00", "abcd", "00000000"}
	for _, v := range valid {
		if !IsValidHexColor(v) {
			t.Errorf("expected valid: %s", v)
		}
	}
	for _, v := range []string{"", "gg", "12345", "#fff", "xyz123"} {
		if IsValidHexColor(v) {
			t.Errorf("expected invalid: %s", v)
		}
	}
}

func TestParseBoolean(t *testing.T) {
	if got := ParseBoolean("true"); got == nil || *got != true {
		t.Error("true should parse true")
	}
	if got := ParseBoolean("FALSE"); got == nil || *got != false {
		t.Error("FALSE should parse false")
	}
	if ParseBoolean("") != nil || ParseBoolean("yes") != nil {
		t.Error("empty/yes should be nil")
	}
}

func TestParseArray(t *testing.T) {
	if len(ParseArray("")) != 0 {
		t.Error("empty should be []")
	}
	if got := ParseArray("a,b,c"); len(got) != 3 || got[1] != "b" {
		t.Errorf("unexpected %v", got)
	}
}

func TestClampValue(t *testing.T) {
	if ClampValue(5, 1, 10) != 5 || ClampValue(-5, 1, 10) != 1 || ClampValue(99, 1, 10) != 10 {
		t.Error("clamp failed")
	}
}

func TestEncodeHTML(t *testing.T) {
	got := EncodeHTML("<hello>&")
	if strings.Contains(got, "<") || strings.Contains(got, ">") {
		t.Errorf("not encoded: %q", got)
	}
}

func TestFlexLayout(t *testing.T) {
	items := FlexLayout([]string{"<a/>", "", "<b/>"}, 10, "", nil)
	if len(items) != 2 {
		t.Fatalf("empty items should be filtered, got %d", len(items))
	}
	if !strings.Contains(items[0], `translate(0, 0)`) || !strings.Contains(items[1], `translate(10, 0)`) {
		t.Errorf("row layout wrong: %v", items)
	}
	col := FlexLayout([]string{"<a/>", "<b/>"}, 5, "column", nil)
	if !strings.Contains(col[1], `translate(0, 5)`) {
		t.Errorf("column layout wrong: %v", col)
	}
}

func TestGetCardColors(t *testing.T) {
	c := GetCardColors("", "", "", "", "", "", "dark", "default")
	if c.TitleColor != "#fff" {
		t.Errorf("dark title should be #fff, got %s", c.TitleColor)
	}
	c2 := GetCardColors("ff0000", "", "", "", "", "", "dark", "default")
	if c2.TitleColor != "#ff0000" {
		t.Errorf("override failed: %s", c2.TitleColor)
	}
}

func TestRenderError(t *testing.T) {
	svg := RenderError("oops <x>", "boom", map[string]string{})
	if !strings.Contains(svg, "Something went wrong") || !strings.Contains(svg, "oops") {
		t.Error("error card missing content")
	}
}

func TestMeasureText(t *testing.T) {
	if MeasureText("", 10) != 0 {
		t.Error("empty should be 0")
	}
	if MeasureText("hello", 10) <= 0 {
		t.Error("should be positive")
	}
}

func TestWrapTextMultiline(t *testing.T) {
	lines := WrapTextMultiline("short desc", 59, 3)
	if len(lines) != 1 {
		t.Errorf("expected 1 line, got %v", lines)
	}
	long := "word1 word2 word3 word4 word5 word6 word7 word8 word9 word10 word11 word12 word13 word14 word15"
	lines = WrapTextMultiline(long, 10, 2)
	if len(lines) != 2 || !strings.HasSuffix(lines[1], "...") {
		t.Errorf("wrap/ellipsis failed: %v", lines)
	}
}

func TestCardRender(t *testing.T) {
	c := NewCard(400, 200, 4.5, GetCardColors("", "", "", "", "", "", "default", "default"), "", "Test", "")
	c.SetHideBorder(false)
	out := c.Render("<text>hi</text>")
	for _, want := range []string{"<svg", "Test", "card-bg", "hi"} {
		if !strings.Contains(out, want) {
			t.Errorf("card missing %q", want)
		}
	}
	c2 := NewCard(400, 200, 4.5, GetCardColors("", "", "", "", "", "", "default", "default"), "", "T", "")
	h0 := c2.Height
	c2.SetHideTitle(true)
	if c2.Height != h0-30 {
		t.Error("hide title should shrink height by 30")
	}
}

func TestCreateProgressNode(t *testing.T) {
	svg := CreateProgressNode(0, 0, 100, "#fff", 50, "#ddd", 100)
	if !strings.Contains(svg, "lang-progress") {
		t.Error("progress node missing class")
	}
	svg = CreateProgressNode(0, 0, 100, "#fff", 500, "#ddd", 0)
	if !strings.Contains(svg, `width="100%"`) {
		t.Errorf("should clamp to 100: %s", svg)
	}
}
