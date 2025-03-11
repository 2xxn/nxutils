# nxutils - a package with various utilities for Go

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
| Subdomains   | GetSubdomains function returns a string slice of subdomains, fetched from crt.sh (with plans to support other websites) |

### web
<!-- | Content recognition | Recognizes content present on the website supporting things such as webmails, DBAs, ASP.NET, gitlab/forgejo, React Apps and much more | -->

| Feature Name | Description |
| ------------ | ----------- |
| React unpacker | A handy tool that unpacks a React website using its map files, returns a VirtualDisk from nxutils/io package |
| Cors Anywhere port checker | A tool that checks open internal ports in cors-anywhere proxies (Blind SSRF) | 