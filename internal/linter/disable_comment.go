package linter

import "strings"

// disabledSet maps absolute line numbers to disabled rule names.
// A nil map value means all rules are disabled on that line.
type disabledSet map[int]map[string]struct{}

func (d disabledSet) isDisabled(line int, ruleName string) bool {
	rules, ok := d[line]
	if !ok {
		return false
	}
	if rules == nil {
		return true
	}
	_, found := rules[ruleName]
	return found
}

func (d disabledSet) addLine(line int, ruleNames []string) {
	if len(ruleNames) == 0 {
		d[line] = nil
		return
	}
	if existing, exists := d[line]; exists && existing == nil {
		return // already all-disabled; nil takes priority
	}
	if _, exists := d[line]; !exists {
		d[line] = make(map[string]struct{})
	}
	for _, r := range ruleNames {
		d[line][r] = struct{}{}
	}
}

// applyBlockState stamps the current block-level disable state onto absLine in set.
func applyBlockState(set disabledSet, absLine int, blockAllDisabled bool, blockRules map[string]struct{}) {
	if blockAllDisabled {
		set[absLine] = nil
		return
	}
	if len(blockRules) == 0 {
		return
	}
	if existing, exists := set[absLine]; exists && existing == nil {
		return // nil (all-disabled) takes priority
	}
	if _, exists := set[absLine]; !exists {
		set[absLine] = make(map[string]struct{})
	}
	for r := range blockRules {
		set[absLine][r] = struct{}{}
	}
}

// parseDisableComments scans lines for gomarklint-disable directives and returns
// a disabledSet mapping absolute line numbers to the set of disabled rules.
// offset is the frontmatter line count used to compute absolute line numbers.
func parseDisableComments(lines []string, offset int) disabledSet {
	set := make(disabledSet)
	blockAllDisabled := false
	blockRules := make(map[string]struct{})

	for i, line := range lines {
		absLine := i + 1 + offset
		directive, ruleNames := parseDirectiveLine(line)

		switch directive {
		case "disable":
			if len(ruleNames) == 0 {
				blockAllDisabled = true
				blockRules = make(map[string]struct{})
			} else {
				for _, r := range ruleNames {
					blockRules[r] = struct{}{}
				}
			}
		case "enable":
			if len(ruleNames) == 0 {
				blockAllDisabled = false
				blockRules = make(map[string]struct{})
			} else {
				for _, r := range ruleNames {
					delete(blockRules, r)
				}
			}
		case "disable-line":
			set.addLine(absLine, ruleNames)
		case "disable-next-line":
			if i+1 < len(lines) {
				set.addLine(absLine+1, ruleNames)
			}
		}

		applyBlockState(set, absLine, blockAllDisabled, blockRules)
	}

	return set
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
