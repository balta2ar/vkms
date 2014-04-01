VKontakte Music Sorter
======================

### About

This application allows you to sync your music from your VK account locally,
sort songs into albums (folders) and sync that album structure back to VK
server.

### Motivation

Because web interface is not convenient enough for this purpose.

### Usage

Single argument - base directory for music is optional.

    go build
    ./vkms [base-music-directory]

### Platforms

It was tested on GNU/Linux only. Due to explicit use of forward slashes
somewhere in the code, it probably won't work on Windows.

### Language

This piece of software is written in [Go](http://golang.org/).

### Bugs

LOTS OF THEM! The code contains mininal amount of error checks so it's very
unreliable (use at your own risk!). However, it looks like it does its job (at
least for me).
