<img align="right" src="https://github.com/h2337/file-hosting/blob/master/rssnix-logo.png?raw=true">

## TOC

* [Demonstration](#demonstration)
* [Installation](#installation)
* [Flags](#flags)
* [Config](#config)

## Demonstration

![Demo](https://raw.githubusercontent.com/h2337/file-hosting/master/rssnix-demo.gif?raw=true)

## Installation

`git clone https://github.com/h2337/rssnix --depth=1 && cd rssnix && go install`

## Flags

`config`
- Opens config file with `$EDITOR`

`update [feed]`
- If [feed] argument is given and is space-delimited list of feeds, then these feeds are updated
- If no [feed] argument is given then all feeds are updated

`open [feed]`
- If [feed] argument is given then the said feed's directory is opened with the configured viewer
- If no [feed] argument is given then the root feeds directory is opened with the configured viewer

`add [feed name] [feed url]`
- Adds a new feed to the config file

## Config

Config file is expected to be at `~/.config/rssnix/config.ini`.

Sample config file:

```
[settings]
viewer = vim
feed_directory = ~/rssnix

[feeds]
CNN-Tech = http://rss.cnn.com/rss/edition_technology.rss
HackerNews = https://news.ycombinator.com/rss
```
(Tip: `ranger` is another great candidate for `viewer`)
