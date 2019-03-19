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
# only type element
parsehtml /path/file.html 'type=element'
parsehtml -html /path/file.html -json '{"type":"element"}'

# match value
parsehtml /path/file.html 'attr.class=preview'
parsehtml -html /path/file.html -json '{"attr":{"class":"preview"}}'

# the null means all match
parsehtml /path/file.html 'data.=a' 'attr.href'
parsehtml -html /path/file.html -json '{"data":"a","attr":{"href":null}}'
```

Use config

```sh
# make template
parsehtml -template > config.json

# edit config.json
$EDITOR config.json

# use
parsehtml -config /pash/config.json -html /path/file.html
```

Regular expression 2

```sh
parsehtml /path/file.html 're2.attr.href=^https?://[\S]+$'
parsehtml -html /path/file.html -json '{"re2":{"attr":{"href":"^https?://[\\S]+$"}}}'
```

If you have `jq`

```sh
# pick one
parsehtml /path/file.html | jq '.[0]'
```

## Installation

```sh
go get github.com/yaeshimo/parsehtml/cmd/parsehtml
```

## License

MIT
