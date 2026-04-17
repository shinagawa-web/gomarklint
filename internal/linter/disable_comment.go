package linter

import "strings"

// disabledSet maps absolute line numbers to disabled rule names.
// An empty (non-nil) slice means all rules are disabled on that line.
// A non-empty slice lists the specific disabled rule names.
type disabledSet map[int][]string

func (d disabledSet) isDisabled(line int, ruleName string) bool {
	rules, ok := d[line]
	if !ok {
		return false
	}
	if len(rules) == 0 {
		return true // all rules disabled
	}
	for _, r := range rules {
		if r == ruleName {
			return true
		}
	}
	return false
}

func (d disabledSet) addLine(line int, ruleNames []string) {
	if len(ruleNames) == 0 {
		d[line] = []string{} // empty non-nil = all rules
		return
	}
	existing, exists := d[line]
	if exists && len(existing) == 0 {
		return // all-disabled takes priority
	}
	d[line] = append(existing, ruleNames...)
}

// applyBlockState stamps the current block-level disable state onto absLine in set.
func applyBlockState(set disabledSet, absLine int, blockAllDisabled bool, blockRules map[string]struct{}) {
	if blockAllDisabled {
		set[absLine] = []string{}
		return
	}
	if len(blockRules) == 0 {
		return
	}
	existing, exists := set[absLine]
	if exists && len(existing) == 0 {
		return // all-disabled takes priority
	}
	names := make([]string, 0, len(blockRules))
	for r := range blockRules {
		names = append(names, r)
	}
	set[absLine] = append(existing, names...)
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
