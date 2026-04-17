package linter

import "strings"

// lineDisable describes which rules are disabled on a single line.
// If allDisabled is true, all rules are disabled except those listed in names (exceptions).
// If allDisabled is false, only the rules listed in names are disabled.
type lineDisable struct {
	allDisabled bool
	names       []string
}

func (ld lineDisable) isRuleDisabled(ruleName string) bool {
	if ld.allDisabled {
		for _, r := range ld.names {
			if r == ruleName {
				return false // exception: explicitly re-enabled
			}
		}
		return true
	}
	for _, r := range ld.names {
		if r == ruleName {
			return true
		}
	}
	return false
}

// disabledSet maps absolute line numbers to their disable state.
type disabledSet map[int]lineDisable

func (d disabledSet) isDisabled(line int, ruleName string) bool {
	ld, ok := d[line]
	if !ok {
		return false
	}
	return ld.isRuleDisabled(ruleName)
}

func (d disabledSet) addLine(line int, ruleNames []string) {
	existing, exists := d[line]
	if exists && existing.allDisabled && len(existing.names) == 0 {
		return // all-disabled (no exceptions) takes priority
	}
	if len(ruleNames) == 0 {
		d[line] = lineDisable{allDisabled: true}
		return
	}
	d[line] = lineDisable{names: append(existing.names, ruleNames...)}
}

// blockState holds the current block-level disable state while scanning lines.
type blockState struct {
	allDisabled bool
	exceptions  []string // re-enabled rules when allDisabled=true
	rules       []string // named disabled rules when allDisabled=false
}

func (bs *blockState) applyTo(set disabledSet, absLine int) {
	if !bs.allDisabled && len(bs.rules) == 0 {
		return
	}
	existing, exists := set[absLine]
	if bs.allDisabled {
		if !exists {
			set[absLine] = lineDisable{allDisabled: true, names: bs.exceptions}
			return
		}
		if existing.allDisabled && len(existing.names) == 0 {
			return // already fully disabled; keep it
		}
		set[absLine] = lineDisable{allDisabled: true, names: bs.exceptions}
		return
	}
	// named-disable mode
	if exists && existing.allDisabled {
		return // all-disabled takes priority
	}
	set[absLine] = lineDisable{names: append(existing.names, bs.rules...)}
}

// parseDisableComments scans lines for gomarklint-disable directives and returns
// a disabledSet mapping absolute line numbers to the set of disabled rules.
// offset is the frontmatter line count used to compute absolute line numbers.
func parseDisableComments(lines []string, offset int) disabledSet {
	set := make(disabledSet)
	var bs blockState

	for i, line := range lines {
		absLine := i + 1 + offset
		directive, ruleNames := parseDirectiveLine(line)

		switch directive {
		case "disable":
			if len(ruleNames) == 0 {
				bs = blockState{allDisabled: true}
			} else {
				bs.rules = append(bs.rules, ruleNames...)
			}
		case "enable":
			if len(ruleNames) == 0 {
				bs = blockState{}
			} else if bs.allDisabled {
				bs.exceptions = append(bs.exceptions, ruleNames...)
			} else {
				bs.rules = removeAll(bs.rules, ruleNames)
			}
		case "disable-line":
			set.addLine(absLine, ruleNames)
		case "disable-next-line":
			if i+1 < len(lines) {
				set.addLine(absLine+1, ruleNames)
			}
		}

		bs.applyTo(set, absLine)
	}

	return set
}

// removeAll returns s with all elements in remove filtered out.
func removeAll(s []string, remove []string) []string {
	result := s[:0]
	for _, v := range s {
		found := false
		for _, r := range remove {
			if v == r {
				found = true
				break
			}
		}
		if !found {
			result = append(result, v)
		}
	}
	return result
}

// parseDirectiveLine extracts the directive keyword and optional rule names
// from a gomarklint HTML comment directive on a line.
// Returns ("", nil) when no valid directive is found.
func parseDirectiveLine(line string) (directive string, ruleNames []string) {
	start := strings.Index(line, "<!--")
	if start == -1 {
		return "", nil
	}
	end := strings.Index(line[start:], "-->")
	if end == -1 {
		return "", nil
	}
	inner := strings.TrimSpace(line[start+4 : start+end])
	const prefix = "gomarklint-"
	if !strings.HasPrefix(inner, prefix) {
		return "", nil
	}
	parts := strings.Fields(inner[len(prefix):])
	if len(parts) == 0 {
		return "", nil
	}
	switch parts[0] {
	case "disable", "enable", "disable-line", "disable-next-line":
		return parts[0], parts[1:]
	default:
		return "", nil
	}
}
