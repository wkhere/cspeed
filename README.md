### cspeed

Measure the speed of ssh connections.

#### usage

```shell
cspeed [-dDT|--duration=DT (default 5s)] HOST [HOSTn...]
```

example:

```shell
% cspeed host1 host2                                                                                                            130
host1: 7670 KBps
host2: 9831 KBps
```

#### how it works

`cspeed` is a wrapper around

```shell
dd if=/dev/urandom | ssh $host 'cat - >/dev/null'
```
with catching of `dd`'s data transfer statistics as a measure good enough
to estimate ssh connection speed.
As simple as that.

Note that `cspeed` runs `ssh` in a subshell, so your existing `.ssh/config` is used.
