# logi-tools

Contains directly usable tools installable via `go get github.com/hirmus/logi-tools/cmd/...`

## logi_loader

Bitstream loader for LOGI Pi FPGA development board

``` text
Usage:
    logi_loader <.bit file>
```

## wb_util

Wishbone util for reading / writing data

``` text
Usage:
    wbutil [-d] [-c X] <address> [write value] .. [write value]
      -c uint
            read count (default 1)
      -d    debug info
```
