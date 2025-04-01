package tblwriter

import (
	"github.com/amterp/color"
	"regexp"
	"strings"
)

var COLOR_ALL = regexp.MustCompile("[\\s\\S]*")

func (t *Table) colorize(s string, c Color) string {
	return t.colorizeWithRegex(s, []ColumnColorMod{NewColumnColorMod(COLOR_ALL, c)})
}

/*
colorizeWithRegex applies color modifications to a string based on a series of regex/color pairs.
The function maintains character-level granularity, meaning each individual character can be colored based
on the highest-priority regex match. The priority of color application is determined by the order in which
the color/regex pairs are provided.

High-level steps:
 1. Initialize a `colorChar` slice where each element represents a character in the original string,
    along with its current color and priority (default is no color and lowest priority).
 2. Apply each regex pattern, updating the color and priority for matching characters.
    Characters are only updated if the new match has a higher priority than any previous match.
 3. Build the final string by grouping consecutive characters with the same color and applying
    the respective color functions in batches to optimize performance.
*/
func (t *Table) colorizeWithRegex(s string, colorMods []ColumnColorMod) string {
	// Skip coloring if it's disabled globally
	if t.forceNoColor {
		return s
	}

	// Each character in the string is represented by a colorChar,
	// which tracks the character, its color, and its priority.
	type colorChar struct {
		char     byte
		color    Color
		priority int
	}

	// Initialize a colorChar slice with each character from the input string,
	// defaulting to no color (Plain) and a priority of -1 (lowest possible).
	colorChars := make([]colorChar, len(s))
	for i := range s {
		colorChars[i] = colorChar{char: s[i], color: Plain, priority: -1}
	}

	// Apply color modifications in the order given.
	// If a character is already colored with a higher-priority color, it won't be overwritten.
	for priority, mod := range colorMods {
		if mod.regex == nil {
			// Skip modifications that have no associated color or regex
			continue
		}
		// Find all matches of the current regex in the string.
		matches := mod.regex.FindAllStringIndex(s, -1)
		for _, match := range matches {
			// For each match, apply the color if the priority is higher than the current one.
			for i := match[0]; i < match[1]; i++ {
				if priority > colorChars[i].priority {
					colorChars[i].color = mod.color
					colorChars[i].priority = priority
				}
			}
		}
	}

	// Prepare the final result by grouping characters with the same color into runs,
	// applying color functions in batches rather than on each individual character.
	var result strings.Builder
	currentColor := Plain
	var currentRun strings.Builder

	// flushRun outputs the accumulated characters in the currentRun using the currentColor.
	flushRun := func() {
		if currentRun.Len() > 0 {
			if currentColor != Plain {
				// Apply the current color to the accumulated characters
				colorFunc := currentColor.toLibColor().SprintFunc()
				result.WriteString(colorFunc(currentRun.String()))
			} else {
				// Output plain text without any coloring
				result.WriteString(currentRun.String())
			}
			// Clear the current run after writing it to the result
			currentRun.Reset()
		}
	}

	// Iterate through each character in colorChars.
	// If the color changes, flush the current run and start a new one with the new color.
	for _, cc := range colorChars {
		if cc.color != currentColor {
			// The color changed, so flush the previous run.
			flushRun()
			// Update the current color to the new color
			currentColor = cc.color
		}
		// Accumulate the current character into the current run.
		currentRun.WriteByte(cc.char)
	}
	// Flush any remaining characters in the last run.
	flushRun()

	// Return the fully built and colored string.
	return result.String()
}

func (t *Table) ToggleColor(enabled bool) {
	color.NoColor = !enabled
}

func (t *Table) SetHeaderColors(colors ...Color) {
	if t.headerMods == nil || len(t.headerMods) == 0 {
		t.headerMods = make([]HeaderMod, len(colors))
	}

	for i, c := range colors {
		t.headerMods[i].color = c
	}
}

func (t *Table) SetHeaderMods(mods ...HeaderMod) {
	t.headerMods = mods
}

func (t *Table) SetColumnMods(mods map[int]ColumnMod) {
	t.columnModsByIdx = mods
}
