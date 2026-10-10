import { App } from '@modelcontextprotocol/ext-apps';
const app = new App({name:'Counter demo',version:'1.0.0'},{});
const count=document.getElementById('count'),button=document.getElementById('increment'),status=document.getElementById('status');
function display(result){ const value=result.structuredContent?.count; if(Number.isInteger(value)) count.textContent=String(value); }
app.ontoolresult=display;
app.onteardown=async()=>{button.disabled=true;status.textContent='Closed';return {};};
button.addEventListener('click',async()=>{
 button.disabled=true;
 const start=performance.now();
 try{display(await app.callServerTool({name:'counter_increment',arguments:{}}));status.textContent=`Ready · ${Math.round(performance.now()-start)} ms`;}
 catch(e){status.textContent=e.message;}
 finally{button.disabled=false;}
});
await app.connect();
status.textContent='Initialized';button.disabled=false;
display(await app.callServerTool({name:'counter_view',arguments:{}}));
