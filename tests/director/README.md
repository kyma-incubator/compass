# Director tests

Director tests consists of acceptance tests.

## Usage

The binary allows you to override some configuration parameters. You can specify the following environment variables:

| ENV                         | Default                         | Description                                       |
|-----------------------------|---------------------------------|---------------------------------------------------|
| ALL_SCOPES                  | ""                              | string with all scopes (permissions) separated by semicolon, which will be used in requests |
| DIRECTOR_URL                | 127.0.0.1:3000                  | The address and port for the Director                           |

To run Director tests with running director and connector, execute:
```
./run.sh
```

To run tests together with Director and Database run:
```bash
make verify 
make clean-up
```
In case of failure 
`clean-up` is required, because in case of test fails, the created network is not deleted.
