# gomarklint demo tape (https://github.com/charmbracelet/vhs)
#
# Regenerate the GIF with:
#   make demo            # builds the binary, then records
#   ./scripts/record-demo.sh
#
# The fixture lives in demo/sample.md — edit it to change what the demo shows.

Output docs/static/demo.gif

Set Shell "bash"
Set FontSize 18
Set Width 1200
Set Height 760
Set Theme "Dracula"
Set Padding 24
Set TypingSpeed 55ms

# Put the freshly built binary on PATH without showing the plumbing.
Hide
Type "export PATH=$PWD:$PATH"
Enter
Type "clear"
Enter
Show

Sleep 800ms

# 1. A Markdown file with a broken anchor link and a couple of structure issues.
Type "cat demo/sample.md"
Sleep 400ms
Enter
Sleep 2.5s

# 2. Lint it — broken links and structure problems caught in milliseconds.
Type "gomarklint demo/sample.md"
Sleep 400ms
Enter
Sleep 4s

Sleep 1.5s
