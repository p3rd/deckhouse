const core = require("@actions/core");

console.log("args", JSON.stringify(process.args));
console.log("env", process.env.INPUT_PRS);

const body = JSON.parse(process.env.INPUT_PRS).map(pr => pr.body).join("\n")

console.log("resulting body", JSON.stringify(body))

core.setOutput("body", body);
