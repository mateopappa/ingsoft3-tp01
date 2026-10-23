import fs from 'node:fs';

const outDir = process.env.OUT_DIR || '/out';
const summary = JSON.parse(fs.readFileSync('coverage/coverage-summary.json', 'utf8'));
const t = summary.total;
const md = `## Frontend coverage (vitest / v8)

| Metric | Value |
|---|---|
| Lines | ${t.lines.pct}% |
| Branches | ${t.branches.pct}% |
| Functions | ${t.functions.pct}% |
| Statements | ${t.statements.pct}% |
| Thresholds | lines ≥ 80, branches ≥ 75 |

Scope: \`src/lib/**/*.js\` (pure logic). DOM controller \`static/js/app.js\` is excluded.
`;
fs.mkdirSync(outDir, { recursive: true });
fs.writeFileSync(`${outDir}/SummaryGithub.md`, md);
fs.writeFileSync('coverage/SummaryGithub.md', md);
console.log(md);
