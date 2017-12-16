# bx-line-notice
Compare and notice price cryptocurrency from bx.in.th

## How to use
- Write the configuration file (config.json)
```json
{
    "BXAPIUrl": "https://bx.in.th/api/",
    "LineAccessToken": "",
    "HTTPTimeout": 10,
    "Delay": 5
}
```
- Run via docker use command

```
docker run -d -v <config.json path>:/root/config.json kusumoto/bx-line-notice
```
