import { journey, loadOptions, stageDuration, maxVUs } from "./common.js";
export { setup, handleSummary } from "./common.js";

export const options = loadOptions("ramp", [
    { duration: stageDuration, target: Math.max(1, Math.round(maxVUs / 20)) },
    { duration: stageDuration, target: Math.max(1, Math.round(maxVUs / 10)) },
    { duration: stageDuration, target: Math.max(1, Math.round(maxVUs / 4)) },
    { duration: stageDuration, target: Math.max(1, Math.round(maxVUs / 2)) },
    { duration: stageDuration, target: maxVUs },
    { duration: stageDuration, target: maxVUs },
    { duration: stageDuration, target: 0 },
]);
export default function() { journey(); }
