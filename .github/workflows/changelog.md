Let's agree on the changelog format

We want
- group changes by module to filter hide modules whihc are not used in cluster
- group changes by type for clarity
- note what will restart, warn or show some other expected cautions, hence to have special field for that
- list resolved issues as well
- include the PR (resolved by workflow)

Let's say, this is PR #150 changelog (milestone v1.42.3). Fields `module`, `type`, `description` are required, others are not.

~~~
```changes
module: module1
type: fix
description: what was fixed in 150
resolves: #aa, #bb
note: Arbitrary description of what will or can restart
---
module: module2
type: feature
description: what was added in 150
resolves: #rr
note: |
  Migration of config values might cause something to restart

```
~~~

Let's say, this is PR #151 changelog (milestone v1.42.3)

~~~
```changes
---
module: module1
type: feature
description: what was added in 151
resolves: #ss
note: |
  Migration of config values might cause something to restart
---
module: module3
type: fix
description: what was fixed in 151
resolves: #oo, #uu
note: Network flap is expected, no longer than 10 seconds
```
~~~

Then the result will look like this in file `CHANGELOG/CHANGELOG-v1.42.3`

```
module1:
  features:
    - description: what was added in 151
      note: |
        Migration of config values might cause something to restart
      pull_request: https://github.com/deckhouse/deckhouse/pulls/151
      resolves:
        - https://github.com/deckhouse/deckhouse/issues/ss
  fixes:
    - description: what was fixed in 150
      note: Arbitrary description of what will or can restart
      pull_request: https://github.com/deckhouse/deckhouse/pulls/150
      resolves:
        - https://github.com/deckhouse/deckhouse/issues/aa
        - https://github.com/deckhouse/deckhouse/issues/bb
module2:
  features:
    - description: what was added in 150
      note: |
        Migration of config values might cause something to restart
      pull_request: https://github.com/deckhouse/deckhouse/pulls/151
      resolves:
        - https://github.com/deckhouse/deckhouse/issues/rr
module3:
  fixes:
    - description: what was fixed in 151
      note: Network flap is expected, no longer than 10 seconds
      pull_request: https://github.com/deckhouse/deckhouse/pulls/150
      resolves:
        - https://github.com/deckhouse/deckhouse/issues/oo
        - https://github.com/deckhouse/deckhouse/issues/uu
```

Для 1.24.16 порлучилось бы так

```changes
module: extended-monitoring
type: fix
description: Catch loop error.
```

```changes
module: prometheus-operator
type: feature
description: Limit scope. It allows users to deploy their own Prometheus operators.
```

```changes
module: dhctl
type: fix
description: Fix proxy command.
```

```changes
module: flant-integration
type: fix
description: Fix empty URLs.
```

```changes
module: flant-integration
type: feature
description: Provide extra metrics.
```

Результат

```
extended-monitoring:
  fixes:
    - description: Catch loop error.
      pull_request: http://github.com/...
prometheus-operator:
  features:
    - description: Limit scope. It allows users to deploy their own Prometheus operators.
      pull_request: http://github.com/...
dhctl:
  fixes:
    - description: Fix proxy command.
      pull_request: http://github.com/...
flant-integration:
  fixes:
    - description: Fix empty URLs.
      pull_request: http://github.com/...
  features:
    - description: Provide extra metrics.
      pull_request: http://github.com/...
```