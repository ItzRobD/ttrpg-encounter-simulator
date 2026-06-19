const fs = require('fs');
const file = 'public/sim_result_response.json';
const content = fs.readFileSync(file, 'utf8');

const allKeys = Array.from(new Set(Array.from(content.matchAll(/"([^"]+)":/g)).map(m => m[1])));
const nonSnake = allKeys.filter(k => /[A-Z]/.test(k)).sort();
console.log('Total keys found:', allKeys.length);
console.log('Non-snake/lowercase keys:');
console.log(nonSnake.join('\n'));
