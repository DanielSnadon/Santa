// Node 24+, Chrome --headless --remote-debugging-port=9224.
import { mkdir, writeFile } from 'node:fs/promises';
import { randomUUID } from 'node:crypto';
const base = process.env.SANTA_TEST_URL ?? 'http://127.0.0.1:5080';
const tabs = await (await fetch('http://127.0.0.1:9224/json')).json();
const tab = tabs.find(t => t.type === 'page');
const ws = new WebSocket(tab.webSocketDebuggerUrl);
await new Promise(resolve => ws.addEventListener('open',resolve,{once:true}));
let next = 0; const pending = new Map();
ws.addEventListener('message',e => { const m=JSON.parse(e.data); if (m.id && pending.has(m.id)) {
    const [resolve,reject]=pending.get(m.id);pending.delete(m.id);m.error?reject(new Error(JSON.stringify(m.error))):resolve(m.result);
}});
function call(method,params={}) { return new Promise((resolve,reject) => { const id=++next;pending.set(id,[resolve,reject]);ws.send(JSON.stringify({id,method,params})); }); }
async function evaluate(expression) { const r=await call('Runtime.evaluate',{expression,returnByValue:true,awaitPromise:true});if(r.exceptionDetails)throw new Error(JSON.stringify(r.exceptionDetails));return r.result.value; }
async function until(predicate) {
    const deadline=Date.now()+15000;
    while(Date.now()<deadline) { if(await predicate())return;await new Promise(r=>setTimeout(r,100)); }
    throw new Error('Browser condition timed out: '+await evaluate('document.body.innerText'));
}
async function navigate(path) { await call('Page.navigate',{url:base+path});await until(()=>evaluate('document.readyState === "complete"')); }
async function form(action,fields,expected) {
    const before=await evaluate('location.href');
    await evaluate(`(() => {const form=document.querySelector('form[action="${action}"]');if(!form)throw new Error('Missing form');const fields=${JSON.stringify(fields)};for(const [key,value]of Object.entries(fields))form.elements.namedItem(key).value=value;form.requestSubmit();})()`);
    await until(async()=>await evaluate(`document.readyState === 'complete' && document.body.innerText.includes(${JSON.stringify(expected)})`));
}
const screenshotDir = process.env.SANTA_SCREENSHOT_DIR ?? '/tmp/santa-ui';
await mkdir(screenshotDir,{recursive:true});
async function reviewPage(name) {
    await evaluate('document.fonts.ready');
    if(!await evaluate(`document.fonts.check('16px "Golos Text"')`))throw new Error('Local font missing');
    for (const width of [360,390,768,1440]) {
        await call('Emulation.setDeviceMetricsOverride',{width,height:900,deviceScaleFactor:1,mobile:width<600});
        const problems = await evaluate(`(() => {
            const problems=[];
            const longTargets=[...document.querySelectorAll('h1,.game-title,.recipient-name,.recipient-email')];
            const originals=longTargets.map(el=>el.textContent);
            longTargets.forEach(el=>el.textContent='ДлинноеИмяБезПробелов'.repeat(12));
            if(document.documentElement.scrollWidth > innerWidth) problems.push('long content overflow');
            longTargets.forEach((el,i)=>el.textContent=originals[i]);
            if(document.documentElement.scrollWidth > innerWidth) problems.push('horizontal overflow');
            for(const input of document.querySelectorAll('input:not([type="hidden"]),textarea')) {
                if(!input.labels?.length) problems.push('unlabelled field: '+input.name);
            }
            for(const image of document.images) if(!image.complete || image.naturalWidth===0) problems.push('broken image');
            return problems;
        })()`);
        if(problems.length)throw new Error(name+' '+width+': '+problems.join(', '));
        if(width===390 || width===1440) {
            const shot=await call('Page.captureScreenshot',{format:'png',captureBeyondViewport:true});
            await writeFile(`${screenshotDir}/${name}-${width}.png`,Buffer.from(shot.data,'base64'));
        }
    }
}
const suffix=randomUUID().slice(0,8);
const emails=[`browser-${suffix}-a@example.test`,`browser-${suffix}-b@example.test`];
const password='Browser-Gift-Password42!';
await call('Page.enable');
await navigate('/');
await call('Page.bringToFront');
await evaluate(`document.querySelector('.skip-link').focus()`);
await call('Input.dispatchKeyEvent',{type:'keyDown',key:'Tab',code:'Tab',windowsVirtualKeyCode:9});
await call('Input.dispatchKeyEvent',{type:'keyUp',key:'Tab',code:'Tab',windowsVirtualKeyCode:9});
if(!await evaluate(`document.activeElement.matches('.brand') && getComputedStyle(document.activeElement).outlineStyle !== 'none'`))throw new Error('Keyboard focus is not visible');
for(let i=0;i<2;i++) {
    await navigate('/register');
    if(i===0)await reviewPage('register');
    await form('/account/register',{email:emails[i],name:`Browser Guest ${i}`,password},'Новая игра');
    if(i===0) { await reviewPage('empty-games'); await form('/account/logout',{},'Создать аккаунт'); await reviewPage('home'); }
}
await form('/actions/create',{title:'Browser game'},'Участники и исключения');
const gamePath=new URL(await evaluate('location.href')).pathname;
const gameId=gamePath.split('/').at(-1);
await reviewPage('draft');
if(!await evaluate(`document.querySelector('form[action$="/publish"] button').disabled`))throw new Error('Draw must be disabled for empty roster');
await form(`/actions/${gameId}/configure`,{emails:emails.join('\n'),exclusions:''},'Сохранено участников: 2');
await form(`/actions/${gameId}/publish`,{},'Вы дарите подарок');
await reviewPage('recipient');
let body=await evaluate('document.body.innerText');
if(!body.includes(emails[0])||body.includes(emails[1]))throw new Error('Recipient privacy violation');
await call('Emulation.setDeviceMetricsOverride',{width:390,height:844,deviceScaleFactor:1,mobile:true});
if(await evaluate('document.documentElement.scrollWidth > innerWidth'))throw new Error('Mobile horizontal overflow');
const shot=await call('Page.captureScreenshot',{format:'png',captureBeyondViewport:true});
await writeFile('/tmp/santa-game-mobile.png',Buffer.from(shot.data,'base64'));
await form('/account/logout',{},'Создать аккаунт');
await navigate('/login');await reviewPage('login');
await form('/account/login',{email:emails[0],password},'Новая игра');
await navigate(gamePath);await until(()=>evaluate(`document.body.innerText.includes('Вы дарите подарок')`));
body=await evaluate('document.body.innerText');
if(!body.includes(emails[1])||body.includes(emails[0]))throw new Error('Second recipient privacy violation');
await call('Emulation.setDeviceMetricsOverride',{width:1280,height:900,deviceScaleFactor:1,mobile:false});
await navigate('/login?error=login_failed');await reviewPage('login-error');
await navigate('/not-found');await reviewPage('not-found');
await navigate('/Error');await reviewPage('error');
console.log('PASS: responsive pages at 360/390/768/1440, labels, images, disabled draw, screenshots;  real browser forms, registration/login/logout, saved roster, draw, two private recipients, mobile width');
ws.close();
