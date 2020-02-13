Daymare
---

`Daymare` is a web log, build with some experiment web technology like native-web-components,
custom-elements, vanilla-javascript, and work in morden browsers only.


## Development
```bash
## install mongodb on macos
$ brew tap mongodb/brew
$ brew install mongodb-community

## or ssh port forward
$ ssh -fN -L 27019:localhost:27017 linode

## hot-reload
$ go get github.com/codegangsta/gin
$ gin -a 8081 run main.go
```
