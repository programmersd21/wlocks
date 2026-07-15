Output demo.gif
Set Width 1100
Set Height 620
Set FontSize 14
Set PlaybackSpeed 1.15

# launch wlocks on a file held open by multiple processes
Type "./demo/wlocks /tmp/wlocks_demo/target.lock"
Enter
Sleep 4s

# navigate through the lock-holding processes
Down
Sleep 450ms
Down
Sleep 450ms
Down
Sleep 450ms
Up
Sleep 450ms

# enter detail view to inspect process metadata
Enter
Sleep 2s

# browse the info tabs
Right
Sleep 500ms
Right
Sleep 500ms
Left
Sleep 500ms
Left
Sleep 500ms

# show kill confirmation prompt, then cancel
Type "K"
Sleep 1s
Escape
Sleep 600ms

# return to list
Escape
Sleep 1s

# fuzzy search
Type "/"
Sleep 200ms
Type "tail"
Sleep 2s
Escape
Sleep 700ms

# cycle sort + reverse
Type "s"
Sleep 500ms
Type "s"
Sleep 500ms
Type "S"
Sleep 700ms

# force refresh
Type "r"
Sleep 2s

# help
Type "?"
Sleep 2s
Escape
Sleep 700ms

# statistics
Type "i"
Sleep 2s
Escape
Sleep 700ms

# command palette
Ctrl+p
Sleep 1s
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms
Enter
Sleep 1s

# theme cycle
Type "T"
Sleep 400ms
Type "T"
Sleep 400ms
Type "T"
Sleep 800ms

# quit
Type "q"
