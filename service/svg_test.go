package service

import (
	"strings"
	"testing"
)

func TestSanitizeSVG_ValidSimple(t *testing.T) {
	input := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><circle cx="50" cy="50" r="40" fill="red"/></svg>`
	result, err := SanitizeSVG(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "<circle") {
		t.Error("expected circle element in output")
	}
	if !strings.Contains(result, "<svg") {
		t.Error("expected svg root in output")
	}
}

func TestSanitizeSVG_AddsDefaults(t *testing.T) {
	input := `<svg><rect x="0" y="0" width="100" height="100" fill="blue"/></svg>`
	result, err := SanitizeSVG(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, `xmlns="http://www.w3.org/2000/svg"`) {
		t.Error("expected xmlns to be added")
	}
	if !strings.Contains(result, `viewBox="0 0 100 100"`) {
		t.Error("expected viewBox to be added")
	}
}

func TestSanitizeSVG_BlocksScript(t *testing.T) {
	input := `<svg xmlns="http://www.w3.org/2000/svg"><script>alert('xss')</script><circle cx="50" cy="50" r="40"/></svg>`
	_, err := SanitizeSVG(input)
	if err == nil {
		t.Fatal("expected error for script element")
	}
	if !strings.Contains(err.Error(), "blocked element") {
		t.Errorf("expected blocked element error, got: %v", err)
	}
}

func TestSanitizeSVG_BlocksForeignObject(t *testing.T) {
	input := `<svg xmlns="http://www.w3.org/2000/svg"><foreignObject><body xmlns="http://www.w3.org/1999/xhtml"><div>hack</div></body></foreignObject></svg>`
	_, err := SanitizeSVG(input)
	if err == nil {
		t.Fatal("expected error for foreignObject element")
	}
}

func TestSanitizeSVG_StripsEventHandlers(t *testing.T) {
	input := `<svg xmlns="http://www.w3.org/2000/svg"><circle cx="50" cy="50" r="40" onclick="alert('xss')" fill="red"/></svg>`
	result, err := SanitizeSVG(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(result, "onclick") {
		t.Error("expected onclick to be stripped")
	}
	if !strings.Contains(result, "fill") {
		t.Error("expected fill attribute to remain")
	}
}

func TestSanitizeSVG_StripsExternalHref(t *testing.T) {
	input := `<svg xmlns="http://www.w3.org/2000/svg"><use href="https://evil.com/sprite.svg#icon"/></svg>`
	result, err := SanitizeSVG(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(result, "evil.com") {
		t.Error("expected external href to be stripped")
	}
}

func TestSanitizeSVG_AllowsFragmentHref(t *testing.T) {
	input := `<svg xmlns="http://www.w3.org/2000/svg"><use href="#mySymbol"/></svg>`
	result, err := SanitizeSVG(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, `href="#mySymbol"`) {
		t.Error("expected fragment href to be kept")
	}
}

func TestSanitizeSVG_SizeLimit(t *testing.T) {
	input := `<svg xmlns="http://www.w3.org/2000/svg">` + strings.Repeat("x", MaxSVGSize+1) + `</svg>`
	_, err := SanitizeSVG(input)
	if err == nil {
		t.Fatal("expected error for oversized SVG")
	}
	if !strings.Contains(err.Error(), "maximum size") {
		t.Errorf("expected size error, got: %v", err)
	}
}

func TestSanitizeSVG_Empty(t *testing.T) {
	_, err := SanitizeSVG("")
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestSanitizeSVG_NoSVGRoot(t *testing.T) {
	input := `<div>not svg</div>`
	_, err := SanitizeSVG(input)
	if err == nil {
		t.Fatal("expected error for non-svg root")
	}
}

func TestSanitizeSVG_StripsDangerousStyle(t *testing.T) {
	input := `<svg xmlns="http://www.w3.org/2000/svg"><rect x="0" y="0" width="100" height="100" style="background: url(javascript:alert(1))"/></svg>`
	result, err := SanitizeSVG(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(result, "javascript") {
		t.Error("expected dangerous style to be stripped")
	}
}

func TestSanitizeSVG_StripsUnknownElements(t *testing.T) {
	input := `<svg xmlns="http://www.w3.org/2000/svg"><div>nope</div><circle cx="50" cy="50" r="40" fill="red"/></svg>`
	result, err := SanitizeSVG(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(result, "<div") {
		t.Error("expected unknown element to be stripped")
	}
	if !strings.Contains(result, "<circle") {
		t.Error("expected circle to remain")
	}
}

func TestSanitizeSVG_ComplexValid(t *testing.T) {
	input := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
		<defs>
			<linearGradient id="grad1">
				<stop offset="0%" stop-color="red"/>
				<stop offset="100%" stop-color="blue"/>
			</linearGradient>
		</defs>
		<rect x="10" y="10" width="80" height="80" fill="url(#grad1)"/>
		<text x="50" y="55" font-size="12" text-anchor="middle" fill="white">Bot</text>
	</svg>`
	result, err := SanitizeSVG(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "linearGradient") {
		t.Error("expected linearGradient in output")
	}
	if !strings.Contains(result, "<text") {
		t.Error("expected text element in output")
	}
}

func TestSanitizeSVG_BlocksIframe(t *testing.T) {
	input := `<svg xmlns="http://www.w3.org/2000/svg"><iframe src="https://evil.com"/></svg>`
	_, err := SanitizeSVG(input)
	if err == nil {
		t.Fatal("expected error for iframe element")
	}
}

func TestSanitizeSVG_PreservesViewBoxCasing(t *testing.T) {
	input := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200"><rect width="200" height="200" fill="red"/></svg>`
	result, err := SanitizeSVG(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, `viewBox="0 0 200 200"`) {
		t.Errorf("expected viewBox with correct casing, got: %s", result)
	}
}

func TestSanitizeSVG_PreservesGradientCasing(t *testing.T) {
	input := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><defs><linearGradient id="g1" gradientUnits="userSpaceOnUse"><stop offset="0" stop-color="red"/><stop offset="1" stop-color="blue"/></linearGradient></defs><rect fill="url(#g1)" width="100" height="100"/></svg>`
	result, err := SanitizeSVG(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "<linearGradient") {
		t.Errorf("expected linearGradient element with correct casing, got: %s", result)
	}
	if !strings.Contains(result, "</linearGradient>") {
		t.Errorf("expected closing linearGradient with correct casing, got: %s", result)
	}
	if !strings.Contains(result, `gradientUnits=`) {
		t.Errorf("expected gradientUnits attribute with correct casing, got: %s", result)
	}
}

func TestSanitizeSVG_BlocksEmbed(t *testing.T) {
	input := `<svg xmlns="http://www.w3.org/2000/svg"><embed src="https://evil.com"/></svg>`
	_, err := SanitizeSVG(input)
	if err == nil {
		t.Fatal("expected error for embed element")
	}
}

func TestSanitizeSVG_BlocksObject(t *testing.T) {
	input := `<svg xmlns="http://www.w3.org/2000/svg"><object data="https://evil.com"/></svg>`
	_, err := SanitizeSVG(input)
	if err == nil {
		t.Fatal("expected error for object element")
	}
}
