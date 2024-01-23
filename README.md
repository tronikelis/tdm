# Tronikel's Dotfile Manager

## Install

```
go install github.com/Tronikelis/tdm@latest
```

## Commands

-   `tdm add`
    -   Loops over synced files and updates them from local files
-   `tdm add [directory]`
    -   Loops over [directory] and copies every file into synced dir
-   `tdm sync`
    -   Loops over synced files and copies them into relevant local files

## Why not just use something that already exists?

First of all, I want to take a shot at this myself, but I found that solutions I tried are complex
and they didn't handle nested `.git` directories

## About .git

Most ppl handle dotfiles in git, and I want to do that as well, so there is a problem when adding nested .git dirs

`tdm` will `.zip` `.git` directories if it encounters one
