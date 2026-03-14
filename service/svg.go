package service

import (
	"encoding/xml"
	"errors"
	"strings"
)

const MaxSVGSize = 10 * 1024 // 10KB

var allowedElements = map[string]bool{
	"svg": true, "g": true, "defs": true, "symbol": true, "use": true,
	"rect": true, "circle": true, "ellipse": true, "line": true,
	"polyline": true, "polygon": true, "path": true,
	"text": true, "tspan": true,
	"lineargradient": true, "radialgradient": true, "stop": true,
	"clippath": true, "mask": true, "title": true, "desc": true,
}

var canonicalElements = map[string]string{
	"svg": "svg", "g": "g", "defs": "defs", "symbol": "symbol", "use": "use",
	"rect": "rect", "circle": "circle", "ellipse": "ellipse", "line": "line",
	"polyline": "polyline", "polygon": "polygon", "path": "path",
	"text": "text", "tspan": "tspan",
	"lineargradient": "linearGradient", "radialgradient": "radialGradient",
	"stop": "stop", "clippath": "clipPath", "mask": "mask",
	"title": "title", "desc": "desc",
}

var blockedElements = map[string]bool{
	"script": true, "foreignobject": true, "iframe": true,
	"object": true, "embed": true,
}

var allowedAttributes = map[string]bool{
	// geometry
	"x": true, "y": true, "x1": true, "y1": true, "x2": true, "y2": true,
	"cx": true, "cy": true, "r": true, "rx": true, "ry": true,
	"width": true, "height": true, "d": true, "points": true,
	// presentation
	"fill": true, "stroke": true, "stroke-width": true, "stroke-linecap": true,
	"stroke-linejoin": true, "stroke-dasharray": true, "stroke-dashoffset": true,
	"stroke-opacity": true, "fill-opacity": true, "opacity": true,
	"fill-rule": true, "clip-rule": true, "font-size": true,
	"font-family": true, "font-weight": true, "text-anchor": true,
	"dominant-baseline": true, "letter-spacing": true,
	// structural
	"viewbox": true, "xmlns": true, "transform": true,
	"id": true, "class": true, "clip-path": true,
	"gradientunits": true, "gradienttransform": true, "offset": true,
	"stop-color": true, "stop-opacity": true, "patternunits": true,
	"maskunits": true, "dx": true, "dy": true,
}

var canonicalAttributes = map[string]string{
	"x": "x", "y": "y", "x1": "x1", "y1": "y1", "x2": "x2", "y2": "y2",
	"cx": "cx", "cy": "cy", "r": "r", "rx": "rx", "ry": "ry",
	"width": "width", "height": "height", "d": "d", "points": "points",
	"fill": "fill", "stroke": "stroke", "stroke-width": "stroke-width",
	"stroke-linecap": "stroke-linecap", "stroke-linejoin": "stroke-linejoin",
	"stroke-dasharray": "stroke-dasharray", "stroke-dashoffset": "stroke-dashoffset",
	"stroke-opacity": "stroke-opacity", "fill-opacity": "fill-opacity", "opacity": "opacity",
	"fill-rule": "fill-rule", "clip-rule": "clip-rule", "font-size": "font-size",
	"font-family": "font-family", "font-weight": "font-weight", "text-anchor": "text-anchor",
	"dominant-baseline": "dominant-baseline", "letter-spacing": "letter-spacing",
	"viewbox": "viewBox", "xmlns": "xmlns", "transform": "transform",
	"id": "id", "class": "class", "clip-path": "clip-path",
	"gradientunits": "gradientUnits", "gradienttransform": "gradientTransform",
	"offset": "offset", "stop-color": "stop-color", "stop-opacity": "stop-opacity",
	"patternunits": "patternUnits", "maskunits": "maskUnits",
	"dx": "dx", "dy": "dy",
}

// SanitizeSVG validates and sanitizes an SVG string using a whitelist approach.
// Returns the sanitized SVG or an error if the input is invalid or contains blocked content.
func SanitizeSVG(input string) (string, error) {
	if len(input) > MaxSVGSize {
		return "", errors.New("SVG exceeds maximum size of 10KB")
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return "", errors.New("SVG content is empty")
	}

	decoder := xml.NewDecoder(input_reader(input))
	decoder.Strict = false
	decoder.AutoClose = xml.HTMLAutoClose

	var out strings.Builder
	foundSVGRoot := false
	depth := 0

	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}

		switch t := token.(type) {
		case xml.StartElement:
			name := strings.ToLower(t.Name.Local)

			if blockedElements[name] {
				return "", errors.New("SVG contains blocked element: " + name)
			}

			if !allowedElements[name] {
				continue
			}

			if depth == 0 && name != "svg" {
				return "", errors.New("root element must be <svg>")
			}
			if name == "svg" && depth == 0 {
				foundSVGRoot = true
			}

			// Filter attributes
			var attrs []xml.Attr
			hasViewBox := false
			hasXmlns := false
			for _, attr := range t.Attr {
				attrName := strings.ToLower(attr.Name.Local)
				attrVal := attr.Value

				// Block event handlers
				if strings.HasPrefix(attrName, "on") {
					continue
				}

				// Block dangerous style values
				if attrName == "style" {
					lower := strings.ToLower(attrVal)
					if strings.Contains(lower, "url(") || strings.Contains(lower, "expression(") || strings.Contains(lower, "javascript:") {
						continue
					}
					attrs = append(attrs, attr)
					continue
				}

				// Filter href/xlink:href — only allow fragment references
				if attrName == "href" || (attr.Name.Space == "xlink" && attrName == "href") {
					if !strings.HasPrefix(attrVal, "#") {
						continue
					}
					attrs = append(attrs, attr)
					continue
				}

				if attrName == "viewbox" {
					hasViewBox = true
				}
				if attrName == "xmlns" || attr.Name.Space == "xmlns" {
					hasXmlns = true
				}

				if allowedAttributes[attrName] {
					attrs = append(attrs, attr)
				}
			}

			// Add defaults for root SVG
			if name == "svg" && depth == 0 {
				if !hasXmlns {
					attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "xmlns"}, Value: "http://www.w3.org/2000/svg"})
				}
				if !hasViewBox {
					attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "viewBox"}, Value: "0 0 100 100"})
				}
			}

			out.WriteString("<")
			out.WriteString(canonicalElements[name])
			for _, attr := range attrs {
				out.WriteString(" ")
				if attr.Name.Space != "" && attr.Name.Space != "xmlns" {
					out.WriteString(attr.Name.Space)
					out.WriteString(":")
				}
				attrLower := strings.ToLower(attr.Name.Local)
				if canonical, ok := canonicalAttributes[attrLower]; ok {
					out.WriteString(canonical)
				} else {
					out.WriteString(attrLower)
				}
				out.WriteString(`="`)
				out.WriteString(escapeAttrValue(attr.Value))
				out.WriteString(`"`)
			}
			out.WriteString(">")
			depth++

		case xml.EndElement:
			name := strings.ToLower(t.Name.Local)
			if allowedElements[name] {
				out.WriteString("</")
				out.WriteString(canonicalElements[name])
				out.WriteString(">")
				depth--
			}

		case xml.CharData:
			if depth > 0 {
				out.WriteString(escapeCharData(string(t)))
			}
		}
	}

	if !foundSVGRoot {
		return "", errors.New("no <svg> root element found")
	}

	return out.String(), nil
}

func input_reader(s string) *strings.Reader {
	return strings.NewReader(s)
}

func escapeAttrValue(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

func escapeCharData(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
