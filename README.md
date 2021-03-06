# planner
Planner is a Go application designed to run on a Raspberry Pi in a wall display for family planning convenience.  The Planner display current weather conditions at your location, as well as forecast for the present day and two days into the future.  It also, displays the Merriam-Webster Word of the Day with pronounciation, part of speech, and definitions, along with your next 10 events in Google Calender.  The background that is displayed is of a random selection of your personal photos that you copy into a defined directory.  You may also define the update frequency of all the data.

## A note concerning background photos:
The most effective photos to chose for use a backgrounds in the Planner are ones that are oriented in the same direction as your display screen.  The display prints in white so photos with a contrasting background are most effective.

## Go Requirements:
Connecting to Google Calender has some special requirements.  Execute the following three commands to obtain the needed packages:

go get golang.org/x/oauth2</br>
go get cloud.google.com/go/compute/metadata</br>
go get google.golang.org/api/calendar/v3</br>

## config.json
json is an easy format for computers to read data.  Small errors can break it, however, so before editing backup the json file and refer to an introductory json syntax reference.  Also, **ALL** lines in the file must remain in place or the planner will break.

JSON | Comments
---- | --------
**"DEBUG":** *true,* | May only be set to true or false.  Currently unused due to lazy programmer.
**"darkSkyKey":** *"",* | This is the key issued to you by darksky.com.  The key shown is a dummy so you must get your own key before the Planner will function. You may obtain a free key at https://darksky.net/dev.
**"latitude":** *"",* | The latitude of your forecast location.
**"longitude":** *"",* | The longitude of your forecast location.
**"excludes":** *"exclude=minutely,hourly,flags",* | This parameter must be used AS IS or the planner will break.
**"weatherURL":** *"https://api.darksky.net/forecast/",* | URL where the weather data is obtained.
**"weatherReloadInterval":** *1,* | This is the frequency, in **HOURS**, with which weather data is updated.  Must be an INTEGER.
**"qotdURL":** *"https://www.quotesdaddy.com/feed",* | Currently unused.
**"qotdReloadInterval":** *12,* | Currently unused.
**"wotdURL":** *"https://www.merriam-webster.com/word-of-the-day",* | URL for Merriam-Webster's **Word of the Day**.
**"wotdReloadInterval":** *12,* | Frequency, in **HOURS**, with which Word of the Day data is refreshed.
**"cssDirectory":** *"./css/planner.css",* | Directory where planner.css is stored.
**"photosDir":** *"./photos",* | Directory where background photos are stored.
**"photoReloadInterval":** *3,* | Frequeny, in **MINUTES**, in which the background photo is changed.
**"timeCheckInterval":** *3,* | Currently unused.  DO NOT REMOVE.
**"HTMLFile":** *"planner.html",* | Path to the *planner.html* file.
**"mwRSS":** *"https://www.merriam-webster.com/wotd/feed/rss2",* | Merriam-Webster Word of the Day URL.
**"mwURL":** *"https://www.dictionaryapi.com/api/v1/references/collegiate/xml/",* | Merriam-Webster Collegiate Dictionary URL
**"mwKEY":** *""* | The key issued to you by Merriam-Webster for use of their API.


Edit **/home/pi/.config/lxsession/LXDE-pi/autostart** so the only lines it contains are:

#@xscreensaver -no-splash # comment this line out to disable screensaver</br>
@xset s 0 0</br>
@xset dpms 0 0 0</br>
@xset s noblank</br>
@xset s noexpose</br>
@chromium-browser --incognito --kiosk file:///home/pi/planner/planner.html</br>
</br>
</br>
</br>
To start the planner program on startup:
There are a lot of ways of doing this, so use whatever suits you.
As the "pi" user, edit the crontab file as follows:

**crontab -e**</br>
</br>
</br>
Go to the end of the file and enter the following on its own line:</br>
</br>
**@reboot /home/pi/planner.planner &**
