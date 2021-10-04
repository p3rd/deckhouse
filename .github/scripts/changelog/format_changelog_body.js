// @ts-check
/*
  pullRequests exmaple:

  [
    {
      "body": "Pull reqeust containing changelog\r\n\r\n```changelog\r\n- module: upmeter\r\n  type: fix\r\n  description: correct group   uptime calculation\r\n  fixes_issues:\r\n    - 13\r\n```\r\n\r\nFollowing is extra comments.",
      "milestone": {
        "number": 2,
        "title": "v1.40.0",
        "description": "",
        "dueOn": null
      },
      "number": 1,
      "state": "MERGED",
      "title": "WIP action draft"
    },
    {
      "body": "body\r\nbody\r\nbody\r\n\r\n```changelog\r\n- module: \"inexisting\"\r\n  type: bug\r\n  description: inexistence was not acknowledged\r\n  resolves: [ \"#6\" ]\r\n  will_restart: null\r\n```",
      "milestone": {
        "number": 2,
        "title": "v1.40.0",
        "description": "",
        "dueOn": null
      },
      "number": 3,
      "state": "MERGED",
      "title": "add two"
    }
  ]
*/

// This function expects an array of pull  requests blonging to single milestone
module.exports = async function (pulls) {
  console.log("passed pull requests", JSON.stringify(pulls, null, 2));

  const changesByModule = collectChangelog(pulls);
  const milestone = pulls.length > 0 ? pulls[0].milestone.title : "";

  console.log({ chlog: changesByModule, milestone });

  const header = `## Changelog ${milestone}`;
  const chlog = JSON.stringify(changesByModule, null, 2);

  const body = [header, chlog].join("\r\n\r\n");
  return body;
};

// pull requests object => changes by modules
function collectChangelog(pullRequests) {
  return pullRequests
    .filter((pr) => pr.state == "MERGED")
    .map(parseChanges)
    .reduce(groupModules, {});
}

// TODO tests on various malformed changelogs
function parseChanges(pr) {
  let rawChanges = "";

  try {
    rawChanges = pr.body.split("```changelog")[1].split("```")[0];
  } catch (e) {
    return `#${pr.number} ${pr.title}`;
  }

  return rawChanges
    .split("---")
    .filter((x) => !!x) // empty strings
    .map((raw) => parseSingleChange(pr, raw));
}

/**
 * @function parseSingleChange parses raw text entry to change object. Multi-line values are not supported.
 * @param {{ url: string; }} pr
 * @param {string} raw

 * Input:
 *
 * `pr`:
 *
 * ```json
 * pr = {
 *   "url": "https://github.com/owner/repo/pulls/151"
 * }
 * ```
 *
 * `raw`:
 *
 * ```change
 * module: module3
 * type: fix
 * description: what was fixed in 151
 * resolves: #16, #32
 * note: Network flap is expected, but no longer than 10 seconds
 * ```
 *
 * Output:
 * ```json
 * {
 *   "module": "module3",
 *   "type": "fix",
 *   "description": "what was fixed in 151",
 *   "note": "Network flap is expected, but no longer than 10 seconds",
 *   "resolves": [
 *     "https://github.com/deckhouse/dekchouse/issues/16",
 *     "https://github.com/deckhouse/dekchouse/issues/32"
 *   ],
 *   "pull_request": "https://github.com/deckhouse/dekchouse/pulls/151"
 * }
 * ```
 *
 */
function parseSingleChange(pr, raw) {
  const lines = raw.split("\n");
  const change = {};

  for (const line of lines) {
    let [k, v] = line.split(":", 1);
    v = v.trim();

    if (!changeFields.has(k)) {
      continue;
    }

    switch (k) {
      // case "resolves":
      //   change[k] = parseIssues(issuesBaseUrl, v);
      //   break;
      default:
        change[k] = v;
    }
  }

  change["pull_request"] = pr.url;
  return change;
}

const changeFields = new Set([
  "module",
  "note",
  "type",
  "description",
  // "resolves",
]);

function parseIssues(baseUrl, v) {
  const nums = v
    .split(",")
    .map((s) => s.trim())
    .filter((x) => !!x);

  if (nums.length == 0) {
    return [];
  }

  return nums
    .map((i) => i.replace("#", ""))
    .map((n) => `${baseUrl}/issues/${n}`);
}

function groupModules(acc, changes) {
  for (const c of changes) {
    addChange(acc, c);
  }
}

function addChange(acc, change) {
  // ensure module key:   { "module": {} }
  acc[change.module] = acc[change.module] || {};
  const mc = acc[change.module];

  // ensure module change list
  // e.g. for fixes: { "module": { "fixes": [] } }
  let list;
  if (change.type == "fix") {
    mc.fixes = mc.fixes || [];
    list = mc.fixes;
  } else if (change.type == "feature") {
    mc.features = mc.features || [];
    list = mc.features;
  }

  // add the change
  list.push({
    description: change.description,
    pull_request: change.pull_request,
    resolves: change.resolves,
    note: change.note,
  });
}
