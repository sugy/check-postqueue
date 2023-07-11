# check-postqueue

## Description

Monitoring queue count in postfix.

## Synopsis

```sh
check-postqueue -w 100 -c 200
```

## Installation

First, build this program.

```sh
go get https://github.com/sugy/check-postqueue
go install
```


Next, you can execute this program :-)

```sh
check-postqueue -w 100 -c 200
```


## Setting for mackerel-agent

If there are no problems in the execution result, add a setting in mackerel-agent.conf .

```conf
[plugin.checks.check-postqueue]
command = ["check-postqueue", "-w", "100", "-c", "200"]
```

## Usage
### Options

```txt
  -w, --warning=  number of messages in queue to generate warning (default: 100)
  -c, --critical= number of messages in queue to generate critical alert ( w < c ) (default: 200)
```

## For more information

Please execute `check-postqueue -h` and you can get command line options.

## other

- Forked [check-mailq](https://github.com/mackerelio/go-check-plugins/tree/master/check-mailq).
