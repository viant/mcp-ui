import {build} from 'esbuild';
import {readFile,writeFile} from 'node:fs/promises';
const result=await build({entryPoints:['app.js'],bundle:true,minify:true,format:'esm',write:false});
const template=await readFile('counter.template.html','utf8');
await writeFile('counter.html',template.replace('<!--APP_SCRIPT-->',`<script type="module">${result.outputFiles[0].text}</script>`));
