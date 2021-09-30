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

  const chlog = collectChangelog(pulls);
  const milestone = pulls.length > 0 ? pulls[0].milestone.title : "";

  console.log({ chlog, milestone });

  const body = [`## Changelog ${milestone}`, chlog].join("\r\n\r\n");

  return body;
};

function collectChangelog(pullRequests) {
  return pullRequests
    .filter((pr) => pr.state == "MERGED")
    .map(parseChangelog)
    .join("\r\n");
}

// TODO tests on various malformed changelogs

/*
TODO changelog format is a matter of discussion


*/
function parseChangelog(pr) {
  try {
    const changelog = pr.body.split("```changelog")[1].split("```")[0];
    return changelog;
  } catch (e) {
    return `#${pr.number} ${pr.title}`;
  }
}
