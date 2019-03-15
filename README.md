# parsehtml

Output json format html nodes to stdout

## Usage

Print usage

```sh
parsehtml -help
```

Output json format html nodes to stdout

```sh
parsehtml /path/file.html
```

Specify filter

```sh
# the null means all match
parsehtml -json '{"attr":{"href":null}}' /path/file.html
```

Use config

```sh
# make template
parsehtml -template > config.json

# edit config.json
$EDITOR config.json

# use
parsehtml -config config.json -html /path/file.html
```

Regular expression 2

```sh
parsehtml /path/file.html '{"re2":{"attr":{"href":"^https?://[\\S]+$"}}}'
```

## Installation

go get

## License

MIT
