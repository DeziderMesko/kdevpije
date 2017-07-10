# kdevpije
Tool aggregating data from Google Calendars, PagerDuty and Google Sheets to show presence of colleagues

## Features:
* in config you can specify aliases for one or more people (horizon or blesk work just fine)
* you can use GoodData abbreviations (MIC, PGR...)
* it uses Levenshtein distance to guess typos in names (marcinisin works just fine ;-)
* google app script regularly update list of employees from google drive employees sheet
* it uses your own google account to access calendars and employees list (no credential rotation needed)

## Bugs, improvements:
* handling of sad PagerDuty API should be more robust, from time to time deserialization of JSON date is failing (null value date)
* decoupling of PD read-only token from binary would be nice (probably creation of some appengine service?)
* better handling of multiple events during day

![Image of example](https://github.com/DeziderMesko/kdevpije/blob/master/kdevpije.png)

