# Screenshot

### Introduction

This application has a wire between GoLang and Objective-C languages.
It needs for to read OSX system events and capture screenshots.

The application provides delivering screenshots to your sftp server.
Just fill fields in the preferences menu and take a screenshot.
Generally, you can do it by combination shift+cmd+4.
After you can put through the link from your buffer by cmd+v.

System requirements:
* OSX 11.5 and higher

### Installation
Download the [latest release](https://github.com/revilon1991/screenshot/releases) and run it on your macOS.

Or build from the source:
```shell
git clone https://github.com/revilon1991/screenshot.git
cd screenshot
make
```
_It's expected that [Make](https://www.gnu.org/software/make/) was installed in your operating system._

### Usage
1. Run `screenshot.app`.
2. Click to 🔲 from menu bar.
3. Fill sftp settings in `Preferences`.
    - _NOTE: Paths must be with end slash._
4. Take a screenshot. Generally, you can do it by hot key combination `cmd + shift + 4`.
5. Put through link your screenshot from the clipboard `cmd + v`.

License
-------

[![license](https://img.shields.io/badge/License-MIT-green.svg?style=flat-square)](./LICENSE)
