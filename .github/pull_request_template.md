
<!--

"Changes" block contains a list of YAML documents. It describes a changelog entry that is formed
automatically.

Fields:

* module      - (required)  Affected module in kebab case, e.g. "node-manager".
* type        - (required)  The change type: only "fix" and "feature" supported.
* description - (optional)  The changelog entry. Omit to use pull request title.
* note        - (optional)  Any notable detail, e.g. expected restarts, downtime, config changes,
                            migrations, etc.

Since the syntax is YAML, `note` may contain multi-line text.

Example:

```changes
module: node-manager
type: fix
description: "Nodes with outdated manifests are no longer provisioned on *InstanceClass update."
note: |
  Expect nodes of "Cloud" type to restart.

  Node checksum calculation is fixes as well as a race condition during
  the machines (MCM) rendering which caused outdated nodes to spawn.
---
module: node-manager
type: feature
description: "Node restarts can be avoided by pinning a checksum to a node group in config values."
note: Recommended to use as last resort.
```

-->

```changes
module: <kebab-case>
type: fix | feature
description: <what effectively changes>
note: <what to expect>
```