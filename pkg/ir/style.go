package ir

// Style represents visual styling properties for nodes and edges.
type Style struct {
	// Visual properties
	Fill         string  `json:"fill,omitempty"`          // Fill color (hex, named, gradient)
	Stroke       string  `json:"stroke,omitempty"`        // Border/line color
	StrokeWidth  int     `json:"stroke_width,omitempty"`  // Border/line width
	StrokeDash   int     `json:"stroke_dash,omitempty"`   // Dash pattern length
	BorderRadius int     `json:"border_radius,omitempty"` // Corner rounding (shapes only)
	Opacity      float64 `json:"opacity,omitempty"`       // Transparency 0.0-1.0

	// Effects
	Shadow       bool `json:"shadow,omitempty"`        // Drop shadow (shapes only)
	ThreeD       bool `json:"3d,omitempty"`            // 3D effect (rectangles/squares only)
	Multiple     bool `json:"multiple,omitempty"`      // Stacked appearance
	DoubleBorder bool `json:"double_border,omitempty"` // Double border (rectangles/ovals)

	// Typography
	Font          string `json:"font,omitempty"`           // Font family
	FontSize      int    `json:"font_size,omitempty"`      // Font size
	FontColor     string `json:"font_color,omitempty"`     // Text color
	Bold          bool   `json:"bold,omitempty"`           // Bold text
	Italic        bool   `json:"italic,omitempty"`         // Italic text
	Underline     bool   `json:"underline,omitempty"`      // Underlined text
	TextTransform string `json:"text_transform,omitempty"` // Text case (uppercase, lowercase, capitalize)

	// Animation (edges only)
	Animated bool `json:"animated,omitempty"` // Animated connection
}

// Merge combines this style with another, with the other style taking precedence.
// Used for cascading styles from containers to children.
func (s Style) Merge(other Style) Style {
	result := s

	if other.Fill != "" {
		result.Fill = other.Fill
	}
	if other.Stroke != "" {
		result.Stroke = other.Stroke
	}
	if other.StrokeWidth != 0 {
		result.StrokeWidth = other.StrokeWidth
	}
	if other.StrokeDash != 0 {
		result.StrokeDash = other.StrokeDash
	}
	if other.BorderRadius != 0 {
		result.BorderRadius = other.BorderRadius
	}
	if other.Opacity != 0 {
		result.Opacity = other.Opacity
	}
	if other.Shadow {
		result.Shadow = other.Shadow
	}
	if other.ThreeD {
		result.ThreeD = other.ThreeD
	}
	if other.Multiple {
		result.Multiple = other.Multiple
	}
	if other.DoubleBorder {
		result.DoubleBorder = other.DoubleBorder
	}
	if other.Font != "" {
		result.Font = other.Font
	}
	if other.FontSize != 0 {
		result.FontSize = other.FontSize
	}
	if other.FontColor != "" {
		result.FontColor = other.FontColor
	}
	if other.Bold {
		result.Bold = other.Bold
	}
	if other.Italic {
		result.Italic = other.Italic
	}
	if other.Underline {
		result.Underline = other.Underline
	}
	if other.TextTransform != "" {
		result.TextTransform = other.TextTransform
	}
	if other.Animated {
		result.Animated = other.Animated
	}

	return result
}
