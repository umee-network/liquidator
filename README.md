# umeeliqd

A basic liquidator tool for Umee's native leverage protocol.

```sh
$ UMEE_LIQUIDATOR_PASSWORD=<KEYRING_PASSPHRASE> umeeliqd --config /path/to/config.toml --log-level info --log-format text
```

Uses a toml config file

```toml
[liquidator]
wait="1m"

[liquidator.select]
reward_denoms=[
  "uumee",
  "ibc/0000000000000000000000000000000000000000000000000000000000000000"
]
repay_denoms=[
  "uumee"
]
```

## Install

To install the `umeeliqd` binary:

```shell
$ make install

## Behavior

When launched, `umeeliqd` performs the following:

- Loads a config file from the supplied path.
- Gets keyring passphrase from an environment variable.
- Starts the main liquidator loop.

The main loop of `umeeliqd` repeats the following steps:

- Retrieves a list of eligible liquidation targets (addresses) from the Umee network:
- Iterates over each target:
  - Selects the preferred reward and repayment denom present in the target.
  - Simulates the result of liquidation using `x/leverage` parameters.
  - Decides whether or not to attempt the liquidation based on estimated rewards.
  - Attempts to execute the liquidation.