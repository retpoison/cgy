cgy
===
A tui youtube rss reader using piped and mpv.


Dependencies
------------
mpv and go.


Installation
------------
    git clone https://github.com/retpoison/cgy.git
    cd cgy
    go build
    ./cgy


Keybindings
-----------
| Key  |    Description |
|:-----|---------------:|
| V, v |----------Videos|
| C, c |--------Channels|
| A, a |-----Add channel|
| R, r |--Refresh videos|
| I, i |-------Instances|
| H, h |------------help|
| Q, q |------------Quit|
| Esc  |------------Back|

Flags
-----
Config path
    cgy -c /home/anon/.config/cgy.json

Piped Instance
    cgy -i https://pipedapi.kavin.rocks

