on appIsRunning(appName)
	tell application "System Events" to (name of processes) contains appName
end appIsRunning

on run
	set theApplication to (name of (info for (path to frontmost application)))
	set theText to ""
	set theURL to ""
	
	if theApplication is "Google Chrome.app" and appIsRunning("Google Chrome") then
		tell application id "com.google.chrome"
			using terms from application "Google Chrome"
				set theText to title of active tab of first window
				set theURL to get URL of active tab of first window
			end using terms from
		end tell
		
	else if theApplication is "Safari.app" and appIsRunning("Safari") then
		tell application id "com.apple.safari"
			using terms from application "Safari"
				set theTab to front document
				set theText to name of theTab
				set theURL to URL of theTab
			end using terms from
		end tell
		
	else if theApplication is "Chromium.app" and appIsRunning("Chromium") then
		tell application "Chromium"
			set theURL to URL of active tab of first window
			set theText to title of active tab of first window
		end tell
		
	else if theApplication is "Firefox.app" and appIsRunning("Firefox") then
		tell application "Firefox"
			activate
			set w to item 1 of window 1
			set theDesc to name of w
		end tell
		tell application "System Events"
			set myApp to name of first application process whose frontmost is true
			if myApp is "Firefox" then
				tell application "System Events"
					keystroke "l" using command down
					delay 0.5
					keystroke "c" using command down
				end tell
				delay 0.5
			end if
			delay 0.5
		end tell
		set theURL to get the clipboard
		
	end if
	
	return {theURL, theText}
	
end run