# bx-line-notice
Compare and notice price cryptocurrency from bx.in.th to line

[![Docker Build Status](https://img.shields.io/docker/build/kusumoto/bx-line-notice.svg)](https://hub.docker.com/r/kusumoto/bx-line-notice/)

## How to use
- Write the configuration file (config.json)
```json
{
    "BXAPIUrl": "https://bx.in.th/api/",
    "LineAccessToken": "",
    "HTTPTimeout": 10,
    "Delay": 5,
    "ReplaceLastData": false
}
```
- Run via docker use command

```
docker run -d -v <config.json path>:/root/config.json kusumoto/bx-line-notice
```
