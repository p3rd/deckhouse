module.exports = async function ({ core }, prs) {
  console.log("passed pull requests", JSON.stringify(prs, null, 2));
  // console.log("input", core.getInput("PULL_REQUESTS", { required: true }))
  // const body = JSON.parse(process.env.INPUT_PRS).map(pr => pr.body).join("\n")
  // console.log("resulting body", JSON.stringify(body))
  core.Output("body", "hren")
};
