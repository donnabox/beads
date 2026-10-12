import { execFileSync } from 'node:child_process';
import { readFileSync } from 'node:fs';
import assert from 'node:assert/strict';
import { markdownAnchors } from '/Users/dbox/repos/wt/vickie/bdp-four-part-editorial/packages/conformance/src/markdown.ts';
const cwd='/Users/dbox/repos/wt/vickie/bdp-four-part-editorial';
const base='cd22ac3e388be3baffaf1d3a4caac7c4bd288502';
const before=execFileSync('git',['show',`${base}:docs/specs/bdp.md`],{cwd,encoding:'utf8'});
const after=readFileSync(`${cwd}/docs/specs/bdp.md`,'utf8');
const oldAnchors=markdownAnchors(before), newAnchors=markdownAnchors(after);
const added=newAnchors.filter(a=>!oldAnchors.some(b=>b.anchor===a.anchor));
for(const old of oldAnchors) assert(newAnchors.some(a=>a.anchor===old.anchor),`lost ${old.anchor}`);
const oldHeads=oldAnchors.filter(a=>a.title!==undefined), newHeads=newAnchors.filter(a=>a.title!==undefined);
const editorial=new Set(['#conformance-profiles-and-reading-guide','#part-i--read','#bead-data-model','#part-ii--readupdate','#part-iii--transactional']);
const body=(source,heads,h)=>source.slice(source.indexOf('\n',h.index)+1,heads.find(a=>a.index>h.index)?.index??source.length).replace(/^ *<a\s+(?:id|name)=["'][^"']+["']\s*><\/a> *$/gm,'').trim();
let checked=0;
for(const old of oldHeads){
 if(editorial.has(old.anchor))continue;
 const current=newHeads.find(a=>a.anchor===old.anchor);
 assert(current,`missing heading ${old.anchor}`);
 assert.equal(body(after,newHeads,current),body(before,oldHeads,old),`changed operative section body: ${old.anchor}`);
 checked++;
}
assert.equal(after.slice(after.indexOf('## Conformance index and informative end matter')),before.slice(before.indexOf('## Conformance index and informative end matter')));
const keywords=text=>Object.fromEntries(['MUST','MAY','SHOULD','SHALL','REQUIRED','OPTIONAL','RECOMMENDED'].map(k=>[k,(text.match(new RegExp(`\\b${k}\\b`,'g'))??[]).length]));
assert.deepEqual(keywords(before),keywords(after));
const tables=s=>s.split('\n').filter(l=>l.startsWith('|')).join('\n');assert.equal(tables(before),tables(after));
const citations=execFileSync('git',['diff',base,'--','packages/conformance/catalog','packages/conformance/development','schemas','packages/protocol/schemas'],{cwd,encoding:'utf8'});assert.equal(citations,'');
const links=[...after.matchAll(/\[[^\]]*\]\((#[^)]+)\)/g)];for(const link of links)assert(newAnchors.some(a=>a.anchor===link[1]),`missing ${link[1]}`);
console.log(JSON.stringify({base,unchangedOperativeSectionBodies:checked,oldAnchorsPreserved:oldAnchors.length,currentAnchors:newAnchors.length,addedAnchors:added.map(a=>a.anchor),internalLinksResolved:links.length,unchangedRequirementKeywords:keywords(after),allTablesUnchanged:true,conformanceEndMatterUnchanged:true,catalogsAndSchemasUnchanged:true},null,2));
