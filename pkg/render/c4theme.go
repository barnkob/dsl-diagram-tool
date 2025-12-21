// Package render provides diagram rendering to various formats.
package render

// C4Classes contains the D2 class definitions for C4 diagram styling.
// These classes follow Structurizr's conventional color scheme:
//   - c4-person: Person actor (dark blue #08427b)
//   - c4-system: Software System (medium blue #1168bd)
//   - c4-container: Container (light blue #438dd5)
//   - c4-component: Component (lightest blue #85bbf0)
//   - c4-external: External System (gray #999999)
//   - c4-external-person: External Person (gray #999999)
//
// Usage in D2:
//
//	customer: Customer {
//	  class: c4-person
//	}
//	banking: Banking System {
//	  class: c4-system
//	}
const C4Classes = `
# C4 Theme Classes (Structurizr color scheme)
classes: {
  c4-person: {
    shape: c4-person
    style.fill: "#08427b"
    style.font-color: "#ffffff"
  }
  c4-system: {
    style.fill: "#1168bd"
    style.font-color: "#ffffff"
  }
  c4-container: {
    style.fill: "#438dd5"
    style.font-color: "#ffffff"
  }
  c4-component: {
    style.fill: "#85bbf0"
    style.font-color: "#000000"
  }
  c4-external: {
    style.fill: "#999999"
    style.font-color: "#ffffff"
  }
  c4-external-person: {
    shape: c4-person
    style.fill: "#999999"
    style.font-color: "#ffffff"
  }
}

`

// ApplyC4Theme prepends C4 class definitions to D2 source code.
// This allows users to apply C4 styling by using class: c4-person, c4-system, etc.
func ApplyC4Theme(source string) string {
	return C4Classes + source
}
