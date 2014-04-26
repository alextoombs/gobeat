gobeat
======
gobeat is a simple application designed to allow you to brag about having beat
someone in a game. Created for bragging about results of ping pong games at
Apcera in a semi-public space.

Created during Gophercon 2014.

# Overview

gobeat is comprised of a command-line tool and a server-side application that
proxies match result reports to Twitter.

The command line tool submits results to Twitter. Or, if you roll that way, you
may simply cURL against the documented API to post a match result.
