# nxutils - a package with various utilities for Go

### This package is strictly for educational purposes only, I'm not responsible for any misuse of this package, pentesting without permission is illegal.

## IMPORTANT
1. This package is still in development, so expect some bugs.
2. This package is not meant to be used in production, it's just a collection of utilities that I've made for myself and will use.

## Features

<!-- ### encoding
todo -->

### io
| Feature Name | Description |
| ------------ | ----------- |
| VirtualDisk  | A single file module that allows you to create a virtual disk stored as a variable, it can create, write and delete files, after you're done with writing data to it, you can export it using SaveAsZip method or CompressAsZip |

### net
| Feature Name | Description |
| ------------ | ----------- |
| Subdomains   | GetSubdomains function returns a string slice of subdomains, fetched from crt.sh (with plans to support other websites / custom methods) |
| Ports | Ports is a module that allows you to scan open port(s) on a target, it supports both TCP and UDP protocols, it can scan a single port or a range of ports with threads |
### web

| Feature Name | Description |
| ------------ | ----------- |
| React unpacker | A handy tool that unpacks a React website using its map files, returns a VirtualDisk from nxutils/io package |
| Cors Anywhere port checker | A tool that checks open internal ports in cors-anywhere proxies (Blind SSRF) | 
| Symfony Profiler downloader | A tool that downloads services' source-code from a website using Symfony Profiler, works only if there's an open dir, returns a VirtualDisk from nxutils/io package |
| (WIP) Content recognition | Recognizes content present on the website supporting things such as webmails, DBAs, ASP.NET, gitlab/forgejo, React Apps and much more |