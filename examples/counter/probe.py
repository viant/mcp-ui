import urllib.request,json,time,pathlib
pathlib.Path('evidence').mkdir(exist_ok=True)
url='http://127.0.0.1:29411/mcp';session=None;seq=0;records=[]
def rpc(method,params={}):
 global session,seq
 seq+=1;headers={'Content-Type':'application/json','Accept':'application/json, text/event-stream'}
 if session:headers['Mcp-Session-Id']=session
 req=urllib.request.Request(url,json.dumps({'jsonrpc':'2.0','id':seq,'method':method,'params':params}).encode(),headers)
 start=time.perf_counter()
 with urllib.request.urlopen(req) as r:
  session=r.headers.get('Mcp-Session-Id',session);raw=r.read().decode()
  if raw.startswith('event:') or raw.startswith('data:'):raw='\n'.join(line[6:] for line in raw.splitlines() if line.startswith('data: '))
  value=json.loads(raw)
 records.append({'method':method,'ms':round((time.perf_counter()-start)*1000,2),'result':value})
 return value
rpc('initialize',{'protocolVersion':'2025-11-25','capabilities':{'extensions':{'io.modelcontextprotocol/ui':{'mimeTypes':['text/html;profile=mcp-app']}}},'clientInfo':{'name':'counter-verifier','version':'1'}})
tools=rpc('tools/list')['result']['tools'];assert {t['name'] for t in tools}=={'counter_view','counter_increment'}
read=rpc('resources/read',{'uri':'ui://counter.demo/counter.html'});assert read['result']['contents'][0]['mimeType']=='text/html;profile=mcp-app'
records[-1]['result']['result']['contents'][0]['text']='[HTML captured separately]'
initial=rpc('tools/call',{'name':'counter_view','arguments':{}})['result']['structuredContent']['count']
for expected in range(initial+1,initial+4):
 value=rpc('tools/call',{'name':'counter_increment','arguments':{}});assert value['result']['structuredContent']['count']==expected
session=None
rpc('initialize',{'protocolVersion':'2025-11-25','capabilities':{},'clientInfo':{'name':'counter-reconnect','version':'1'}})
assert rpc('tools/call',{'name':'counter_view','arguments':{}})['result']['structuredContent']['count']==initial+3
pathlib.Path('evidence/protocol.json').write_text(json.dumps(records,indent=2));print(json.dumps([{'method':r['method'],'ms':r['ms']} for r in records],indent=2))
