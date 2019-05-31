# CD project

containerM helps you to automatically deploy your docker containers into google cloud

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

```
docker-compose up -d
```

### Prerequisites

For use

 - docker ^18.06
 - docker-compose ^1.20

For develop

 - go ^1.10

### Installing

```
➜  ~ go get -u -v github.com/jonyhy96/containerM
```

```
➜  containerM git:(master) go run main.go
2019/05/31 10:54:34 main.go:45: Server started at port:8888
```

## Running the tests

### Params

| param | required | e.g. |
| :-------- | :-----: | :----: |
| pk     | true |   VE9LRU4=     |
| image  | true |   registry.domain.com/pk-go/master   |
| env    | false |  \\{\"KEY\":\"VALUE\"\\}  |
| ports  | false |  8080,9090 | 

```
➜  ~ curl -v "localhost:8888/hooks?pk=VE9LRU4=&image=registry.domain.com/pk-go/master&env=\{\"KEY\":\"VALUE"\}&ports=9999,8989"
```

### Coding style

[CODEFMT](https://github.com/golang/go/wiki/CodeReviewComments)

## Deployment

```
make init
docker-compose up -d
```

## Contributing

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct, and the process for submitting pull requests to us.

## Versioning

We use [SemVer](http://semver.org/) for versioning. For the versions available, see the [tags on this repository](https://gitlab.domain.com/golang/containerM/tags). 

## Authors

* **HAO YUN** - *Initial work* - [haoyun](https://github.com/jonyhy96)

See also the list of [contributors](CONTRIBUTORS.md) who participated in this project.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details

## Acknowledgments

* nothing