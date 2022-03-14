# umeeliqd

A basic liquidator tool for Umee's native leverage protocol.

```sh
export UMEE_LIQUIDATOR_PASSWORD={keyring_password}
$ umeeliqd --config {path/to/config.toml}
```

Uses a toml config file

```toml
[log]
level="info"
format="text"
```

### Behavior

When launched, umeeliqd does the following
- loads a config file from the supplied path
- gets keyring passwrod from environment variable
- starts its main loop

The main loop of umeeliqd repeats the following steps
- gets a list of eligible liquidation targets (addresses) from umeed
- iterates over each target:
  - selects preferred reward and repayment denom present on target
  - simulates the result of liquidation using umeed `x/leverage` parameters
  - decides whether or not to attempt the liquidation based on estimated rewards
  - executes the liquidation