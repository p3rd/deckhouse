<!-- Describe the changes so they will be added to a release changelog. Find examples below. -->

```changes
module: <kebab-case>
type: fix | feature
description: <what effectively changes>
note: <what to expect>
```

<!--

"Changes" block contains a list of YAML documents. It describes a changelog entry that is collected
to a release changelog.

Fields:

module      - Required. Affected module in kebab case, e.g. "node-manager".
type        - Required. The change type: only "fix" and "feature" supported.
description - Optional. The changelog entry. Omit to use pull request title.
note        - Optional. Any notable detail, e.g. expected restarts, downtime, config changes, migrations, etc.

Since the syntax is YAML, `note` may contain multi-line text.

There can be multiple docs in single `changes` block, and multiple `changes` blocks in the PR body.

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
module: cloud-provider-aws
type: feature
description: "Node restarts can be avoided by pinning a checksum to a node group in config values."
note: Recommended to use as a last resort.
```
-->
