#!/usr/bin/env bash
set -euo pipefail

# Prompt for macOS password via GUI (hidden input) and print it to stdout for sudo.
osascript <<'APPLESCRIPT'
Tell application "System Events"
  Activate
  set dlg to display dialog "Docker Desktop install needs admin access.\n\nEnter your macOS password:" default answer "" with hidden answer buttons {"OK"} default button "OK"
  set pw to text returned of dlg
end tell
return pw
APPLESCRIPT

