const $=s=>document.querySelector(s);
const SV=p=>'<svg viewBox="0 0 24 24">'+p+'</svg>';
const ICON={
  ai:SV('<rect x="4" y="8" width="16" height="12" rx="3"/><path d="M12 4v4M9 13h.01M15 13h.01M9.5 16.5h5"/>'),
  reply:SV('<path d="M9 14L4 9l5-5M4 9h10a6 6 0 0 1 6 6v5"/>'),
  react:SV('<circle cx="12" cy="12" r="9"/><path d="M8.5 14s1.3 2 3.5 2 3.5-2 3.5-2M9 9.5h.01M15 9.5h.01"/>'),
  chev:SV('<path d="M6 9l6 6 6-6"/>'),
  lock:SV('<rect x="5" y="11" width="14" height="10" rx="2"/><path d="M8 11V7a4 4 0 0 1 8 0v4"/>'),
  edit:SV('<path d="M4 20h4L19 9l-4-4L4 16v4zM13.5 6.5l4 4"/>'),
  del:SV('<path d="M4 7h16M10 11v6M14 11v6M6 7l1 13h10l1-13M9 7V4h6v3"/>'),
};
const REACTS=[
  ['agree','согласен',SV('<path d="M7 10v11H3V10h4zm0 0l4-7a2 2 0 0 1 3 2l-1 5h6a2 2 0 0 1 2 2.3l-1.3 7A2 2 0 0 1 17.7 21H7"/>')],
  ['disagree','не согласен',SV('<g transform="rotate(180 12 12)"><path d="M7 10v11H3V10h4zm0 0l4-7a2 2 0 0 1 3 2l-1 5h6a2 2 0 0 1 2 2.3l-1.3 7A2 2 0 0 1 17.7 21H7"/></g>')],
  ['seen','видел',SV('<path d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7S2 12 2 12z"/><circle cx="12" cy="12" r="3"/>')],
  ['question','непонятно',SV('<circle cx="12" cy="12" r="9"/><path d="M9.3 9a2.8 2.8 0 0 1 5.4 1c0 2-2.7 2.6-2.7 4M12 17h.01"/>')],
  ['done','сделано',SV('<path d="M20 6L9 17l-5-5"/>')],
  ['smile','смешно',SV('<circle cx="12" cy="12" r="9"/><path d="M8 14s1.5 2.5 4 2.5 4-2.5 4-2.5M9 9.5h.01M15 9.5h.01"/>')],
  ['kiss','поцелуйчик',SV('<path d="M2.5 12c2.5-4 5.5-5.5 7.5-3.5 1.2-1 2.8-1 4 0 2-2 5-.5 7.5 3.5-2.5 4-6 6-9.5 6s-7-2-9.5-6z"/><path d="M2.5 12c3.5 1 6.5 1 9.5 0 3 1 6 1 9.5 0"/>')],
  ['thanks','спасибо',SV('<path d="M12 20s-7-4.4-9-8.6A4.8 4.8 0 0 1 12 6.5a4.8 4.8 0 0 1 9 4.9C19 15.6 12 20 12 20z"/>')],
  ['fire','огонь',SV('<path d="M12 3s5.5 4.5 5.5 10a5.5 5.5 0 0 1-11 0c0-2.6 1.6-4.4 1.6-4.4s.4 2.2 2.1 2.8C9.8 7.6 12 3 12 3z"/>')],
  ['deal','договорились',SV('<path d="M2 11l4-4 4 1.5L13 7l4.5 1L22 12"/><path d="M6 14l3.5 3.5a1.4 1.4 0 0 0 2-2l1 1a1.4 1.4 0 0 0 2-2l1 1a1.4 1.4 0 0 0 2-2L13 7"/><path d="M2 11l4 3"/>')],
  ['look','иду смотреть',SV('<circle cx="10.5" cy="10.5" r="6.5"/><path d="M15.5 15.5L21 21"/>')],
  ['cool','прикольно',SV('<path d="M12 3l1.6 5.4L19 10l-5.4 1.6L12 17l-1.6-5.4L5 10l5.4-1.6z"/><path d="M19 15l.7 2.3L22 18l-2.3.7L19 21l-.7-2.3L16 18l2.3-.7z"/>')],
];
const HUES=[28,210,150,280,340,190,50,0,250,110];
let KEY='',ME='',got=0,TY='',LV='',WORK=[],_ty=0,ADMIN=false,ADMINS=[],PEOPLE=[],TOPICS=[],T=null,last=0,gen=0,msgs={},R={},RX={},replyTo=0,quoteText='',toSet=new Set(),priv=false,files=[],seen={},shut={},notified={},goneSeen={},HOL={},live=false,atCur=null,_j=0,HEAD=0,RS=new Set(),SKIP=null;
try{KEY=localStorage.getItem('svyazKey')||'';seen=JSON.parse(localStorage.getItem('svyazSeen')||'{}');shut=JSON.parse(localStorage.getItem('svyazShut')||'{}');goneSeen=JSON.parse(localStorage.getItem('svyazGone')||'{}')}catch(e){}
const store=(k,v)=>{try{localStorage.setItem(k,JSON.stringify(v))}catch(e){}};
const esc=s=>String(s).replace(/[&<>"]/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;'}[c]));
const hue=n=>HUES[Math.max(0,PEOPLE.indexOf(n))%HUES.length];
const genit=n=>n.split(' ').map(w=>{const l=w.slice(-1),p=w.slice(-2,-1);if(w.length>4&&/ая$/.test(w))return w.slice(0,-2)+'ой';if(w.length>4&&/(ова|ева|ёва|ина|ына)$/.test(w)&&!/^(Ирина|Марина|Алина|Полина|Карина|Галина|Регина|Кристина|Валентина|Екатерина|Ангелина|Арина|Зарина|Альбина|Каролина|Эвелина|Мальвина|Сабина|Устина|Акулина)$/.test(w))return w.slice(0,-1)+'ой';if(l==='а'||l==='я')return w.slice(0,-1)+('гкхжчшщ'.includes(p)||l==='я'?'и':'ы');if(l==='й'||l==='ь')return w.slice(0,-1)+'я';if('еёиоуыэю'.includes(l))return w;return w+'а'}).join(' ');
const label=(n,ai,s)=>ai?'ИИ '+genit(n)+(s?' · '+s:''):n;
const fmt=s=>new Date(s).toLocaleString('ru',{day:'numeric',month:'short',hour:'2-digit',minute:'2-digit'});
window.addEventListener('unhandledrejection',e=>{if(e.reason&&e.reason.status)$('#err2').textContent=e.reason.message});
async function api(q,body){
  if(body&&!(body instanceof FormData))body=new URLSearchParams(body);
  const r=await fetch('api.php?'+new URLSearchParams(q),{method:body?'POST':'GET',headers:{'X-Key':KEY},body});
  const j=r.status===204?null:await r.json();if(!r.ok){const e=new Error(j&&j.error||r.status);e.status=r.status;throw e}return j;
}
async function login(){
  try{const me=await api({a:'me'});ME=me.name;ADMIN=me.admin;ADMINS=me.admins||[];$('#invbtn').hidden=!me.admin}catch(e){$('#err').textContent=KEY?'Ключ не подошёл':'';return}
  try{localStorage.setItem('svyazKey',KEY)}catch(e){}
  PEOPLE=await api({a:'people'});
  $('#login').style.display='none';$('#app').style.display='grid';$('#who').textContent=ME;
  if('Notification' in window&&Notification.permission==='default')document.addEventListener('click',()=>Notification.requestPermission(),{once:true});
  await loadTopics();setInterval(loadTopics,15000);
  loadHol();setInterval(loadHol,30*60000);stars();
}
function goneNote(g){
  const k=g.id+'|'+g.ts;if(goneSeen[k])return;goneSeen[k]=1;store('svyazGone',goneSeen);
  const txt=g.by+' убрал(а) тебя из темы «'+g.title+'»'+(g.why?', причина: '+g.why:'');
  const d=document.createElement('div');d.className='toast';d.innerHTML='<span></span><button type="button" class="icon" title="Закрыть"><svg viewBox="0 0 24 24"><path d="M6 6l12 12M18 6L6 18"/></svg></button>';
  d.firstChild.textContent=txt;d.lastChild.onclick=()=>d.remove();$('#toasts').append(d);
  if(document.hidden&&'Notification' in window&&Notification.permission==='granted')new Notification('Связь',{body:txt,tag:'svyazgone'+g.id});
  if(T&&T.id===g.id)goneHere(txt);
}
function goneHere(txt){gen++;const d=document.createElement('div');d.className='sys';d.textContent=txt;$('#msgs').append(d);$('#compose').hidden=true;$('#head').hidden=true;T=null}
async function loadTopics(){
  TOPICS=await api({a:'topics'});
  api({a:'gone'}).then(l=>l.forEach(goneNote)).catch(()=>{});
  const groups={};
  for(const t of TOPICS){
    const o=t.archived?[]:t.members.filter(n=>n!==ME),k=t.archived?'Архив':o.join(', ')||'Только я';
    if(t.archived&&shut[k]===undefined)shut[k]=true;
    (groups[k]=groups[k]||{k,o,ts:[],last:''}).ts.push(t);
    if(t.last>groups[k].last)groups[k].last=t.last;
    if(T&&t.id===T.id&&!document.hidden)seen[t.id]=t.count;
    const un=t.count-(seen[t.id]||0);
    if(un>0&&(document.hidden||!T||T.id!==t.id)&&notified[t.id]!==t.count&&seen[t.id]!==undefined&&'Notification' in window&&Notification.permission==='granted'){
      new Notification('Связь: '+t.title,{body:'Новых: '+un,tag:'svyaz'+t.id});
    }
    notified[t.id]=t.count;
    if(seen[t.id]===undefined)seen[t.id]=t.count;
  }
  store('svyazSeen',seen);
  const box=$('#topics');box.innerHTML='';let total=0;
  for(const g of Object.values(groups).sort((a,b)=>(a.k==='Архив')-(b.k==='Архив')||b.last.localeCompare(a.last))){
    const gd=document.createElement('div');gd.className='grp'+(shut[g.k]?' shut':'');
    let gun=0;
    const h=document.createElement('div');h.className='gh';
    h.innerHTML=(g.k==='Архив'?[]:g.o.length?g.o:[ME]).map(n=>'<span class="dot" style="--h:'+hue(n)+'"></span>').join('')+'<span class="names">'+esc(g.k)+'</span><span class="badge" hidden></span>'+ICON.chev.replace('<svg','<svg class="chev"');
    h.onclick=()=>{shut[g.k]=!shut[g.k];store('svyazShut',shut);gd.classList.toggle('shut')};
    gd.append(h);
    for(const t of g.ts.sort((a,b)=>b.last.localeCompare(a.last))){
      const un=Math.max(0,t.count-(seen[t.id]||0));gun+=un;
      const d=document.createElement('div');d.className='topic'+(T&&t.id===T.id?' on':'');
      d.innerHTML='<div class="tt"><div>'+esc(t.title)+'</div><small>№'+t.id+' · '+fmt(t.last)+'</small></div>'+(pend(t).length?'<span class="atb">@</span>':'')+(un?'<span class="badge">'+un+'</span>':'');
      d.onclick=()=>open(t);d.oncontextmenu=e=>{e.preventDefault();topicMenu(t,e.clientX,e.clientY)};gd.append(d);
    }
    if(gun){const b=h.querySelector('.badge');b.hidden=false;b.textContent=gun}
    total+=gun;box.append(gd);
  }
  document.title=(total?'('+total+') ':'')+'Связь';renderAt();
  if(T){const t=TOPICS.find(x=>x.id===T.id);if(t&&(t.members.length!==T.members.length||!!t.archived!==!!T.archived||t.title!==T.title||JSON.stringify(t.ais)!==JSON.stringify(T.ais))){T=t;renderHead()}}
}
function topicMenu(t,x,y){
  closeTopicMenu();
  const m=document.createElement('div');m.className='menu';m.id='tmenu';
  m.style.cssText='position:fixed;right:auto;top:'+y+'px;left:'+x+'px';
  m.innerHTML='<div>'+(t.archived?'Вернуть из архива':'В архив')+'</div>';
  m.firstChild.onclick=async()=>{closeTopicMenu();await api({a:'archive',t:t.id,on:t.archived?0:1});loadTopics()};
  document.body.append(m);
  const r=m.getBoundingClientRect();
  if(r.right>innerWidth)m.style.left=Math.max(4,innerWidth-r.width-4)+'px';
  if(r.bottom>innerHeight)m.style.top=Math.max(4,y-r.height)+'px';
}
function closeTopicMenu(){const m=$('#tmenu');if(m)m.remove()}
document.addEventListener('mousedown',e=>{if(!e.target.closest('#tmenu'))closeTopicMenu()});
document.addEventListener('keydown',e=>{if(e.key==='Escape')closeTopicMenu()});
document.addEventListener('scroll',()=>{const m=$('#tmenu');if(m&&!m.classList.contains('kickbox'))m.remove()},true);
let AIST={};
function paintTy(){const l=TY.split(',');document.querySelectorAll('.ty').forEach(e=>e.classList.toggle('on',l.includes(e.dataset.n)))}
function paintWork(){
  document.querySelectorAll('#msgs .wk').forEach(e=>e.remove());
  for(const w of WORK){
    const d=document.getElementById('m'+w.on);if(!d||!d.querySelector('.rs'))continue;
    const x=document.createElement('div');x.className='wk';
    x.innerHTML=esc(label(w.n,true))+(w.ans?' ответил: <span class="ref" data-go="'+w.ans+'">№'+w.ans+'</span>':' взял<span class="ty on"><i></i><i></i><i></i></span>');
    d.querySelector('.rs').before(x);
  }
}
function paintUse(u){
  const tk=n=>n>=1e6?(n/1e6).toFixed(1).replace('.',',')+' млн':n>=1e3?Math.round(n/1e3)+' тыс.':String(n);
  const dt=s=>s?new Date(s).toLocaleString('ru-RU',{day:'2-digit',month:'2-digit',hour:'2-digit',minute:'2-digit'}).replace(',',''):'?';
  const sum=x=>tk(x.tok)+' токенов · $'+x.usd.toFixed(2);
  const was=$('#use details')?.open,h=u.flatMap(x=>x.hist.map(p=>Object.assign({n:x.n},p))).sort((a,b)=>a.till<b.till?1:-1);
  $('#use').hidden=!u.length;
  $('#use').innerHTML=esc('Расход ИИ с подключения: '+u.map(x=>x.n+(x.since?' (с '+dt(x.since)+')':'')+' — '+sum(x)).join(' · '))
    +(h.length?'<details'+(was?' open':'')+'><summary>Прошлые сессии</summary>'+h.map(p=>'<div>'+esc(p.n+' · '+dt(p.since)+' – '+dt(p.till)+' — '+sum(p))+'</div>').join('')+'</details>':'');
}
function paintAist(){
  const t={on:'на связи',think:'думает',off:'не на связи'};
  document.querySelectorAll('.aist').forEach(e=>{const s=AIST[e.dataset.n]||'';e.className='aist '+s;e.title=s?label(e.dataset.n,true)+' '+t[s]:''});
}
function renderHead(){
  $('#ttl').textContent=T.title;$('#archbtn').textContent=T.archived?'Вернуть из архива':'В архив';
  const kick=ADMIN||T.by===ME;
  $('#members').innerHTML=T.members.map(n=>'<span class="chip"><span class="dot" style="--h:'+hue(n)+'"></span>'+esc(n)+'<span class="ty" data-n="'+esc(n)+'"><i></i><i></i><i></i></span><span class="aist" data-n="'+esc(n)+'">ИИ<i></i><i></i><i></i></span>'+(kick&&n!==ME?'<button type="button" class="kick" data-kick="'+esc(n)+'" title="Убрать из темы"><svg viewBox="0 0 24 24"><path d="M6 6l12 12M18 6L6 18"/></svg></button>':'')+'</span>').join('')
    +(kick&&T.members.length>2?'<button type="button" class="chip kickall" data-kick="*">Убрать всех</button>':'');
  paintAist();paintTy();
  $('#toRow').innerHTML='Кому:';
  const pc=document.createElement('span');pc.className='chip priv'+(priv?' on':'');pc.innerHTML=ICON.lock+'Шёпотом своему ИИ';
  pc.onclick=()=>{priv=!priv;renderHead()};$('#toRow').append(pc);
  $('#toRow').classList.toggle('privon',priv);
  for(const n of T.members)for(const [ai,s] of [[false,''],[true,''],...((T.ais||{})[n]||[]).map(s=>[true,s])]){
    const l=label(n,ai,s);if(l===ME)continue;
    const c=document.createElement('span');c.className='chip'+(toSet.has(l)?' on':'');c.style.setProperty('--h',hue(n));
    c.innerHTML=(ai?ICON.ai:'')+esc(l);
    c.onclick=()=>{toSet.has(l)?toSet.delete(l):toSet.add(l);c.classList.toggle('on')};
    $('#toRow').append(c);
  }
  const s=document.createElement('span');s.textContent=priv?'видите только ты и '+label(ME,true):'не выбрано — всем';s.style.marginLeft='4px';$('#toRow').append(s);
}
async function loadHol(){try{HOL=await api({a:'holiday'})}catch(e){HOL={}}renderHol()}
function renderHol(){
  const h=HOL.today||HOL.tomorrow,bar=$('#hol'),root=document.documentElement;
  document.body.classList.toggle('hol',!!HOL.today);
  if(h)root.style.setProperty('--hh',h.hue);
  const key='svyazHolX'+(h?h.from+(HOL.today?'t':'m'):'');
  let off=false;try{off=!!localStorage.getItem(key)}catch(e){}
  bar.hidden=!h||!T||off;if(bar.hidden)return;
  const fmtD=d=>new Date(d+'T12:00').toLocaleDateString('ru-RU',{day:'numeric',month:'long'});
  bar.querySelector('b').textContent=(HOL.today?'Сегодня — ':'Завтра — ')+h.name;
  bar.querySelector('span').textContent=HOL.today?(h.from!==h.to?'до '+fmtD(h.to):(h.kind==='проф'?'профессиональный праздник':'')):'';
  $('#holgo').hidden=!HOL.today;
  $('#holx').onclick=()=>{try{localStorage.setItem(key,'1')}catch(e){}bar.hidden=true};
}
const STAR='<svg viewBox="0 0 24 24" fill="currentColor" stroke="none"><path d="M12 2l1.8 6.2L20 10l-6.2 1.8L12 18l-1.8-6.2L4 10l6.2-1.8z"/><path d="M19 15l.8 2.2L22 18l-2.2.8L19 21l-.8-2.2L16 18l2.2-.8z" opacity=".6"/></svg>';
function confetti(hue){
  if(matchMedia('(prefers-reduced-motion: reduce)').matches)return;
  for(let i=0;i<40;i++){
    const c=document.createElement('i');c.className='conf';
    c.style.cssText='left:'+(Math.random()*100)+'vw;background:hsl('+(hue+Math.random()*80-40)+' 80% 60%);animation-delay:'+(Math.random()*.4)+'s;--dx:'+(Math.random()*160-80)+'px;--r:'+(Math.random()*720-360)+'deg';
    document.body.append(c);setTimeout(()=>c.remove(),3000);
  }
}
function pw(el){
  if(matchMedia('(prefers-reduced-motion: reduce)').matches)return;
  const b=el.getBoundingClientRect(),x=b.left+b.width/2,y=b.top+b.height/2;
  for(let i=0;i<36;i++){
    const c=document.createElement('i'),a=Math.random()*Math.PI*2,d=70+Math.random()*170;c.className='conf pw';
    c.style.cssText='left:'+x+'px;top:'+y+'px;background:hsl('+(Math.random()*360)+' 85% 62%);--dx:'+Math.cos(a)*d+'px;--dy:'+Math.sin(a)*d+'px;--r:'+(Math.random()*720-360)+'deg';
    document.body.append(c);setTimeout(()=>c.remove(),1000);
  }
}
$('#holgo').onclick=async()=>{
  const v=$('#text').value.trim(),fd=new FormData();fd.append('text',v);fd.append('card','1');
  try{await api({a:'post',t:T.id},fd)}catch(er){$('#err2').textContent=er.message;return}
  $('#text').value='';fit();confetti(HOL.today.hue);
};
const _u=s=>new TextDecoder().decode(Uint8Array.from(atob(s),c=>c.charCodeAt(0))),_h=s=>{let x=5381;for(const c of s.toLowerCase().replace(/[^а-яёa-z]/g,''))x=(x*33+c.charCodeAt(0))>>>0;return x};
function _q(c,html,ms){if(document.querySelector('.'+c))return;const d=document.createElement('div');d.className=c;d.innerHTML=html;document.body.append(d);setTimeout(()=>d.classList.add('o'),ms-800);setTimeout(()=>d.remove(),ms)}
let _t=[];
$('#side header').addEventListener('mousemove',e=>{const l=$('#logo'),r=l.getBoundingClientRect(),dx=e.clientX-(r.left+r.width/2),dy=e.clientY-(r.top+r.height/2),d=Math.hypot(dx,dy);l.style.transform=d<70&&d>0?'translate('+(dx/d*Math.min(5,70-d)).toFixed(1)+'px,'+(dy/d*Math.min(5,70-d)).toFixed(1)+'px)':''});
$('#side header').addEventListener('mouseleave',()=>{$('#logo').style.transform=''});
$('#logo').onclick=()=>{const n=Date.now();_t=_t.filter(x=>n-x<2500).concat(n);if(_t.length<8)return;_t=[];
  _q('q8','<svg viewBox="0 0 200 100"><path d="M100 50C80 15 30 15 30 50s50 35 70 0 50-35 70 0-50 35-70 0"/></svg><b>'+_u('0LHQtdGB0LrQvtC90LXRh9C90LDRjyDRgdCy0Y/Qt9GM')+'</b><span>'+_u('0LjQtNC10Y8g4oCUIEBzZXlzaGVlIMK3INGB0LTQtdC70LDQu9C4INCc0LDRiNCwINC4INC10ZEg0JjQmA==')+'</span>',4600)};
const stars=()=>{const h=new Date().getHours();document.body.classList.toggle('n7',dark()&&(h>=23||h<5));};
setInterval(stars,600000);
let _k='';document.addEventListener('keydown',e=>{_k=(_k+String.fromCharCode(e.keyCode)).slice(-10);if(_k==="&&((%'%'BA"){_k='';document.body.classList.add('r8');setTimeout(()=>document.body.classList.remove('r8'),10000)}});
const _m=v=>{if(_h(v)===4229350210)_q('m3','<svg viewBox="0 0 24 24"><path d="M20 14.5A8 8 0 1 1 9.5 4a6.5 6.5 0 0 0 10.5 10.5z"/></svg>',9000)};
let holTaps=[];
$('#hol b').onclick=()=>{const n=Date.now();holTaps=holTaps.filter(x=>n-x<1500).concat(n);if(holTaps.length>=5){holTaps=[];confetti(HOL.today?HOL.today.hue:HOL.tomorrow.hue)}};
const pend=t=>t.calls||[];
function renderAt(){
  const b=$('#atbtn'),n=TOPICS.reduce((s,t)=>s+pend(t).length,0);
  const ct=T&&TOPICS.find(t=>t.id===T.id),w=new Set(ct?pend(ct):[]);
  document.querySelectorAll('#msgs .m').forEach(d=>d.classList.toggle('waiting',w.has(+d.id.slice(1))));
  b.hidden=!n||!T;if(b.hidden)return;
  b.querySelector('span').textContent=n;
}
$('#atbtn').onclick=async()=>{
  const all=[];TOPICS.filter(t=>T&&t.id===T.id).concat(TOPICS.filter(t=>!T||t.id!==T.id)).forEach(t=>pend(t).forEach(id=>all.push({t,id})));
  if(!all.length)return;
  const i=atCur?all.findIndex(x=>x.t.id===atCur.t.id&&x.id===atCur.id):-1,nx=all[(i+1)%all.length];atCur=nx;
  if(!T||T.id!==nx.t.id)open(nx.t);
  let el=null;for(let k=0;k<40&&!(el=document.getElementById('m'+nx.id));k++)await new Promise(r=>setTimeout(r,100));
  if(el){el.scrollIntoView({behavior:'smooth',block:'center'});el.classList.add('flash');setTimeout(()=>el.classList.remove('flash'),1500)}
};
function quote(id,part){
  const m=msgs[id];if(!m)return '';
  return '<div class="q'+(part?' part':'')+'" data-go="'+id+'"><b>'+esc(m.from||m.who)+':</b> '+esc(m.gone?'сообщение удалено':part||(m.text||'').slice(0,200)||'[файл]')+'</div>';
}
function fileUrl(f){return 'api.php?'+new URLSearchParams({a:'file',t:T.id,f})}
function build(m){
  const d=document.createElement('div');d.id='m'+m.id;
  if(m.gone){d.className='sys';d.textContent=label(m.who,m.ai,m.sub)+': сообщение удалено · '+fmt(m.ts);return d}
  const MB=(m.who===ME&&!m.ai?'<button type="button" class="icon" data-edit title="Изменить">'+ICON.edit+'</button><button type="button" class="icon" data-del title="Удалить">'+ICON.del+'</button>':'');
  d.className='m'+(m.ai?' ai':'')+(m.who===ME?' mine':'')+(m.priv?' priv':'')+(m.card?' card':'');if(m.card)d.style.setProperty('--ch',m.card.hue);d.style.setProperty('--h',hue(m.who));
  d.innerHTML='<div class="by">'+(m.ai?ICON.ai:'')+'<b>'+esc(m.from||label(m.who,m.ai))+'</b>'
    +(m.to?'<span class="to">→ '+esc(m.to.join(', '))+'</span>':'')+(m.priv?'<span class="to pv">'+ICON.lock+'шёпот</span>':'')+'<span class="to wt">ждёт решения</span>'+(m.edited?'<span>изменено</span>':'')+'<span>'+fmt(m.ts)+' · №'+m.id+'</span></div>'
    +(m.reply?quote(m.reply,m.quote):'')+(m.card?'<div class="cardh">'+STAR+'<span>'+esc(m.card.name)+'</span></div>':'')+'<div class="txt"></div>'
    +'<div class="files"></div>'
    +'<div class="rs"></div><div class="act">'+'<button type="button" class="icon" data-reply title="Ответить">'+ICON.reply+'</button><button type="button" class="icon" data-pal title="Реакция">'+ICON.react+'</button>'+MB+'</div>'+'<div class="act bot">'+'<button type="button" class="icon" data-reply title="Ответить">'+ICON.reply+'</button><button type="button" class="icon" data-pal title="Реакция">'+ICON.react+'</button>'+MB+'</div>';
  if(/^\**\s*Резюме:/.test(m.text||'')){d.classList.add('fold');const f=document.createElement('button');f.type='button';f.className='foldbtn';f.textContent='Резюме — развернуть';f.onclick=()=>{f.textContent=d.classList.toggle('fold')?'Резюме — развернуть':'Резюме — свернуть'};d.querySelector('.txt').before(f)}
  const tx=d.querySelector('.txt'),fl=m.files||[],used=new Set(),parts=(m.text||'').split(/\[фото (\d+)\]/);
  parts.forEach((p,i)=>{
    if(i%2===0){if(i>0)p=p.replace(/^\n/,'');if(i<parts.length-1)p=p.replace(/\n$/,'');if(p)rich(tx,p);return}
    const f=fl[p-1];if(!f||used.has(f)){tx.append('[фото '+p+']');return}
    used.add(f);const s=document.createElement('div');s.className='files';s.innerHTML=media(f);tx.append(s);
  });
  const rest=fl.filter(f=>!used.has(f)),fd=d.querySelector(':scope>.files');
  if(rest.length)fd.innerHTML=rest.map(media).join('');else fd.remove();
  return d;
}
function redraw(id){
  const el=document.getElementById('m'+id);if(!el)return;
  const nd=build(msgs[id]);el.replaceWith(nd);
  if(!msgs[id].gone){if(nd.offsetHeight>240)nd.classList.add('long');renderReacts(id)}
  document.querySelectorAll('.q[data-go="'+id+'"]').forEach(q=>{const p=q.closest('.m'),pm=p&&msgs[+p.id.slice(1)];q.outerHTML=quote(id,pm&&pm.quote)});
  renderAt();paintWork();
}
function addMsg(m){
  const box=$('#msgs');
  if(m.react){
    const k=m.who+'|'+(m.ai?1:0)+'|'+m.react,rk=m.on+'|'+k;if(SKIP&&SKIP.has(rk))return;RS.add(rk);R[m.on]=R[m.on]||{};
    if(m.off)delete R[m.on][k];else R[m.on][k]=m;
    if(RX[m.on])renderReacts(RX[m.on]);else{RX[m.id]=m.on;renderReacts(m.on)}
    last=m.id;return;
  }
  if(m.edit||m.del){
    last=m.id;const o=msgs[m.edit||m.del],el=document.getElementById('m'+(m.edit||m.del));if(!o)return;
    if(m.del)msgs[m.del]={id:o.id,who:o.who,ai:o.ai,gone:true,ts:o.ts};else{o.text=m.text;o.edited=m.ts}
    if(el&&!el.querySelector('.ed'))redraw(o.id);
    return;
  }
  msgs[m.id]=m;last=m.id;
  if(m.sys){const d=document.createElement('div');d.id='m'+m.id;if(/позвал|убрал\(а\) из/.test(m.text))_j=0;d.className='sys';d.textContent=m.text+' · '+fmt(m.ts);box.append(d);return}
  const d=build(m);
  if(m.gone){box.append(d);return}
  if([100,500,1000].includes(++_j)&&new Set(T.members).size>1){d.querySelector('.by').insertAdjacentHTML('beforeend','<span class="to k5">'+_u('0Y7QsdC40LvQtdC5IMK3IA==')+_j+'</span>');if(live)confetti(50)}
  box.append(d);if(live)renderAt();if(d.offsetHeight>240)d.classList.add('long');renderReacts(m.id);if(m.card&&live&&m.who!==ME)confetti(m.card.hue);
}
function rich(el,t){const L=t.split('\n'),row=x=>/^\s*\|.*\|\s*$/.test(x),cells=x=>x.trim().slice(1,-1).split('|').map(c=>c.trim());let buf=[];
  const flush=()=>{if(buf.length){refs(el,buf.join('\n'));buf=[]}};
  for(let i=0;i<L.length;i++){
    if(row(L[i])&&L[i+1]!==undefined&&/^\s*\|[\s:|-]+\|\s*$/.test(L[i+1])&&L[i+1].includes('-')){
      flush();const tb=document.createElement('table');tb.className='tbl';const tr=(c,tag)=>{const r=tb.insertRow();c.forEach(x=>{const d=document.createElement(tag);refs(d,x);r.append(d)})};
      tr(cells(L[i]),'th');i+=2;while(i<L.length&&row(L[i]))tr(cells(L[i++]),'td');i--;
      const w=document.createElement('div');w.className='tblw';w.append(tb);el.append(w);if(L[i+1]==='')i++;continue}
    buf.push(L[i]);}
  flush()}
function refs(el,t){t.split(EMOR).forEach((y,j)=>{if(j%2){const s=document.createElement('span');s.className='emo';s.textContent=y.slice(1,-1);el.append(s);return}y.split(/(№\d+)/).forEach((x,i)=>{if(i%2){const r=document.createElement('span');r.className='ref';r.dataset.go=x.slice(1);r.textContent=x;el.append(r)}else if(x)el.append(x)})})}
function media(f){return /\.(mp4|webm|mov)$/i.test(f.f)?'<video controls preload="metadata" src="'+esc(fileUrl(f.f))+'"></video>':'<a href="'+esc(fileUrl(f.f))+'" target="_blank"><img loading="lazy" alt="'+esc(f.name)+'" src="'+esc(fileUrl(f.f))+'"></a>'}
function renderReacts(id){
  const d=document.getElementById('m'+id);if(!d||!d.querySelector('.rs'))return;
  const by={};for(const e of Object.values(R[id]||{}))(by[e.react]=by[e.react]||[]).push(e);
  d.querySelector('.rs').innerHTML=REACTS.filter(r=>by[r[0]]).map(([k,name,ic])=>{
    const es=by[k],mine=es.some(e=>e.who===ME&&!e.ai);
    const rr={};for(const e of es)for(const x of Object.values(R[e.id]||{}))(rr[x.react]=rr[x.react]||[]).push(label(x.who,x.ai,x.sub)+' ответил(а) '+label(e.who,e.ai,e.sub));
    const bad=REACTS.filter(r=>rr[r[0]]).slice(0,2).map(([q,qn,qi])=>'<i class="rr" title="'+esc(qn+': '+rr[q].join('; '))+'">'+qi+'</i>').join('');
    return '<span class="r'+(mine?' me':'')+'" data-r="'+k+'" title="'+esc(name+': '+es.map(e=>label(e.who,e.ai,e.sub)+' · '+fmt(e.ts)).join('; '))+'">'+ic+es.length+bad+'</span>';
  }).join('');
}
async function react(id,k){
  const mine=R[id]&&R[id][ME+'|0|'+k];
  await api({a:'react',t:T.id,on:id,r:k,off:mine?1:0});
}
$('#msgs').onclick=e=>{
  const im=e.target.closest('.files a');
  if(im){e.preventDefault();$('#lb img').src=im.href;$('#lb').hidden=false;return}
  const j=e.target.closest('.k5');if(j){pw(j);return}
  const q=e.target.closest('[data-go]');
  if(q){const t=document.getElementById('m'+q.dataset.go);if(t){t.scrollIntoView({behavior:'smooth',block:'center'});t.classList.add('flash');setTimeout(()=>t.classList.remove('flash'),1200)}return}
  const m=e.target.closest('.m');if(!m)return;const id=+m.id.slice(1);
  if(e.target.closest('[data-reply]')){startReply(id);return}
  if(e.target.closest('[data-edit]')){startEdit(m,id);return}
  const dl=e.target.closest('[data-del]');if(dl){if(!dl.classList.contains('arm')||Date.now()-dl.dataset.t<400){if(!dl.classList.contains('arm')){dl.classList.add('arm');dl.dataset.t=Date.now();setTimeout(()=>dl.classList.remove('arm'),3000)}return}api({a:'del',t:T.id,on:id}).catch(er=>{$('#err2').textContent=er.message});return}
  const r=e.target.closest('[data-r]');if(r){react(id,r.dataset.r);return}
  const pk=e.target.closest('.pal [data-k]');if(pk){react(+pk.closest('.pal').dataset.on||id,pk.dataset.k);closePal();return}
  if(e.target.closest('[data-pal]')){
    const open=m.querySelector('.pal');closePal();if(open)return;
    const p=document.createElement('div');p.className='pal'+(e.target.closest('.act.bot')?' frombot':m.getBoundingClientRect().bottom>$('#msgs').getBoundingClientRect().bottom-90?' up':'');
    p.innerHTML=REACTS.map(([k,name,ic])=>'<button type="button" class="icon" data-k="'+k+'">'+ic+name+'</button>').join('');
    m.append(p);m.classList.add('picking');
  }
};
function rePal(e){
  const r=e.target.closest('.rs [data-r]'),m=r&&r.closest('.m');if(!m)return;e.preventDefault();
  const es=Object.values(R[+m.id.slice(1)]||{}).filter(x=>x.react===r.dataset.r),t=es.filter(x=>!(x.who===ME&&!x.ai)).pop()||es.pop();if(!t)return;
  closePal();const p=document.createElement('div');p.className='pal'+(m.getBoundingClientRect().bottom>$('#msgs').getBoundingClientRect().bottom-90?' up':'');p.dataset.on=t.id;
  p.innerHTML=REACTS.map(([k,name,ic])=>'<button type="button" class="icon" data-k="'+k+'">'+ic+name+'</button>').join('');
  m.append(p);m.classList.add('picking');
}
$('#msgs').addEventListener('contextmenu',rePal);
let _lp=0;$('#msgs').addEventListener('pointerdown',e=>{if(e.pointerType!=='touch'||!e.target.closest('.rs [data-r]'))return;_lp=setTimeout(()=>{_lp=-1;rePal(e)},500)});
$('#msgs').addEventListener('pointerup',()=>{if(_lp>0)clearTimeout(_lp)});
$('#msgs').addEventListener('click',e=>{if(_lp===-1){_lp=0;e.stopImmediatePropagation()}},true);
function startEdit(m,id){
  if(m.querySelector('.ed'))return;
  const tx=m.querySelector('.txt'),ta=document.createElement('textarea'),bar=document.createElement('div');
  ta.className='ed';ta.value=msgs[id].text||'';bar.className='edbar';
  bar.innerHTML='<button type="button" class="ghost">Отмена</button><button type="button">Сохранить</button>';
  tx.hidden=true;tx.after(ta,bar);ta.focus();
  const done=()=>{ta.remove();bar.remove();redraw(id)};
  const save=async()=>{const v=ta.value.trim();if(!v||v===msgs[id].text){done();return}if(bar.lastChild.disabled)return;bar.lastChild.disabled=true;try{await api({a:'edit',t:T.id,on:id},{text:v});msgs[id].text=v;done()}catch(er){bar.lastChild.disabled=false;$('#err2').textContent=er.message}};
  bar.firstChild.onclick=done;bar.lastChild.onclick=save;
  ta.onkeydown=e=>{if(e.key==='Escape')done();if(e.key==='Enter'&&(e.ctrlKey||e.metaKey))save()};
}
function closePal(){document.querySelectorAll('.pal').forEach(p=>{p.parentNode.classList.remove('picking');p.remove()})}
document.addEventListener('click',e=>{if(!e.target.closest('.m'))closePal();if(!e.target.closest('.drop'))$('#addmenu').hidden=true});
const drafts={};
function saveDraft(){if(T)drafts[T.id]={text:$('#text').value,files:files.slice()}}
async function open(t){
  saveDraft();if(rec)rec.stop();
  T=t;AIST={};last=0;got=0;HEAD=0;RS=new Set();TY='';LV='';WORK=[];_j=0;msgs={};R={};RX={};replyTo=0;quoteText='';toSet=new Set();priv=false;files=[];$('#att').innerHTML='';
  const dr=drafts[t.id]||{text:'',files:[]};$('#text').value=dr.text;fit();files=dr.files.slice();renderAtt();$('#replyBar').hidden=true;$('#err2').textContent='';
  const g=++gen;$('#msgs').innerHTML='';$('#head').hidden=false;$('#compose').hidden=false;
  renderHead();renderHol();$('#app').classList.add('open');
  seen[t.id]=t.count;store('svyazSeen',seen);loadTopics();
  let first=true;
  while(g===gen){
    try{
      const r=await api(Object.assign({a:'read',t:T.id,since:last,wait:first?0:20,lv:LV},first?{tail:200}:{}));
      if(g!==gen)break;
      AIST=r.ai_state||{};paintAist();TY=(r.typing||[]).join(',');paintTy();LV=r.lv||'';WORK=r.work||[];paintUse(r.usage||[]);
      const box=$('#msgs'),down=first||box.scrollHeight-box.scrollTop-box.clientHeight<80;
      live=!first;if(first){got=r.skip||0;_j=r.j||0}got+=r.messages.length;r.messages.forEach(addMsg);if(first&&r.skip)moreBtn(r);paintWork();if(first)renderAt();if(down)box.scrollTop=box.scrollHeight;first=false;
      if(!document.hidden){seen[T.id]=got;store('svyazSeen',seen)}
    }catch(e){if(e.status===410){if(g===gen)goneHere(e.message.split('. Слежку')[0]);break}await new Promise(r=>setTimeout(r,5000))}
  }
}
function moreBtn(r){
  HEAD=r.messages.length?r.messages[0].id:HEAD;
  const b=document.createElement('button');b.type='button';b.className='ghost';b.id='more';b.textContent='Раньше';b.onclick=earlier;$('#msgs').prepend(b);
}
async function earlier(){
  const b=$('#more'),g=gen;b.disabled=true;
  try{
    const r=await api({a:'read',t:T.id,since:0,tail:200,before:HEAD});if(g!==gen)return;
    const box=$('#msgs'),h=box.scrollHeight,top=box.scrollTop,keep=document.createDocumentFragment();b.remove();
    while(box.firstChild)keep.append(box.firstChild);
    const l=last,j=_j,lv=live;live=false;_j=r.j||0;SKIP=new Set(RS);
    try{r.messages.forEach(addMsg)}finally{SKIP=null;last=l;_j=j;live=lv}
    box.append(keep);if(r.skip)moreBtn(r);box.scrollTop=top+box.scrollHeight-h;
  }catch(e){b.disabled=false;$('#err2').textContent=e.message}
}
try{const w=+localStorage.getItem('svyazSide');if(w)document.documentElement.style.setProperty('--sw',w+'px')}catch(e){}
$('#sgrip').onpointerdown=e=>{
  const g=e.target;g.setPointerCapture(e.pointerId);g.classList.add('on');document.body.style.userSelect='none';
  const mv=ev=>document.documentElement.style.setProperty('--sw',Math.min(560,Math.max(220,ev.clientX))+'px');
  g.onpointermove=mv;
  g.onpointerup=()=>{g.onpointermove=null;g.classList.remove('on');document.body.style.userSelect='';try{localStorage.setItem('svyazSide',parseInt(getComputedStyle(document.documentElement).getPropertyValue('--sw')))}catch(e){}};
};
const dark=()=>(document.documentElement.dataset.theme||(matchMedia('(prefers-color-scheme: dark)').matches?'dark':'light'))==='dark';
const themeIcon=()=>{$('#themebtn').innerHTML=dark()?'<svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="4"/><path d="M12 2v2M12 20v2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M2 12h2M20 12h2M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4"/></svg>':'<svg viewBox="0 0 24 24"><path d="M20 14.5A8 8 0 1 1 9.5 4a6.5 6.5 0 0 0 10.5 10.5z"/></svg>'};
themeIcon();
$('#themebtn').onclick=()=>{const t=dark()?'light':'dark';document.documentElement.dataset.theme=t;try{localStorage.setItem('svyazTheme',t)}catch(e){}themeIcon();stars()};
$('#go').onclick=()=>{KEY=$('#key').value.trim();login()};
$('#key').onkeydown=e=>{if(e.key==='Enter')$('#go').click()};
document.addEventListener('visibilitychange',()=>{if(document.hidden&&rec)rec.stop();if(!document.hidden&&T){seen[T.id]=got;store('svyazSeen',seen);loadTopics()}});
$('#members').onclick=async e=>{
  const b=e.target.closest('[data-kick]');if(!b)return;
  const n=b.dataset.kick,r=b.getBoundingClientRect();closeTopicMenu();
  const m=document.createElement('div');m.className='menu kickbox';m.id='tmenu';
  m.style.cssText='position:fixed;left:auto;right:'+Math.max(8,innerWidth-r.right)+'px;top:'+(r.bottom+6)+'px';
  m.innerHTML='<b></b><div class="pick">'+['Вопрос решён','Не по адресу','Позвали по ошибке'].map(x=>'<span class="chip">'+x+'</span>').join('')+'</div><input placeholder="Своя причина"><div class="btns"><button type="button" class="ghost">Без причины</button><button type="button">Убрать</button></div>';
  m.querySelector('b').textContent=n==='*'?'Убрать всех, кроме тебя и '+label(ME,true):'Убрать: '+n+' и '+label(n,true);
  const inp=m.querySelector('input');
  m.querySelectorAll('.chip').forEach(c=>c.onclick=()=>{inp.value=c.textContent;inp.focus()});
  const go=async why=>{closeTopicMenu();await api(Object.assign(n==='*'?{a:'kick',t:T.id,all:1}:{a:'kick',t:T.id,who:n},why?{why}:{}));T.members=T.members.filter(x=>n==='*'?x===ME:x!==n);renderHead();loadTopics()};
  const [no,yes]=m.querySelectorAll('.btns button');no.onclick=()=>go('');yes.onclick=()=>go(inp.value.trim());
  inp.onkeydown=e=>{if(e.key==='Enter')yes.click()};
  document.body.append(m);inp.focus();
};
$('#goto').onkeydown=e=>{if(e.key!=='Enter')return;const el=document.getElementById('m'+parseInt(e.target.value.replace(/\D/g,'')));if(!el){e.target.classList.add('bad');setTimeout(()=>e.target.classList.remove('bad'),800);return}e.target.value='';e.target.blur();el.scrollIntoView({block:'center'});el.classList.add('flash');setTimeout(()=>el.classList.remove('flash'),1500)};
$('#taskbtn').onclick=async()=>{
  const ks=await api({a:'tasks',t:T.id}),box=$('#tasklist');
  const one=k=>'<details class="task"><summary><b>'+k.n+'.</b> '+esc(k.title||'без названия')+'<small>'+(k.state==='todo'&&k.prio?'приоритет '+k.prio+' · ':'')+fmt(k.upd||'')+'</small></summary>'
    +(k.info?'<div class="info">'+esc(k.info)+'</div>':'')
    +'<ol>'+k.log.map(e=>'<li>'+fmt(e.ts)+' · '+esc(e.by)+': '+esc(e.what)+(e.msg?' · <a data-tgo="'+e.msg+'">№'+e.msg+'</a>':'')+'</li>').join('')+'</ol></details>';
  const sec=(h,l)=>'<div><h4>'+h+'</h4>'+(l.length?l.map(one).join(''):'<div class="none">пусто</div>')+'</div>';
  const now=ks.filter(k=>k.state==='now'),todo=ks.filter(k=>k.state==='todo').sort((a,b)=>b.prio-a.prio),done=ks.filter(k=>k.state==='done');
  box.innerHTML=ks.length?sec('В работе',now)+sec('Не в работе',todo)+(done.length?'<details class="arch"><summary>Архив · '+done.length+'</summary>'+done.map(one).join('')+'</details>':''):'<div class="none">Задач пока нет — их заводит ИИ темы, когда в переписке появляется задача.</div>';
  $('#taskmodal').hidden=false;
};
$('#tasklist').onclick=e=>{
  const a=e.target.closest('[data-tgo]');if(!a)return;
  const m=document.getElementById('m'+a.dataset.tgo);if(!m)return;
  $('#taskmodal').hidden=true;m.scrollIntoView({behavior:'smooth',block:'center'});m.classList.add('flash');setTimeout(()=>m.classList.remove('flash'),1200);
};
$('#archbtn').onclick=async()=>{const on=!T.archived;await api({a:'archive',t:T.id,on:on?1:0});T.archived=on;renderHead();loadTopics()};
$('#ren').onclick=()=>{
  const h=$('#ttl');h.contentEditable=true;h.focus();getSelection().selectAllChildren(h);
};
$('#ttl').onkeydown=e=>{
  if(e.key==='Enter'){e.preventDefault();$('#ttl').blur()}
  if(e.key==='Escape'){$('#ttl').textContent=T.title;$('#ttl').blur()}
};
$('#ttl').onblur=async()=>{
  const h=$('#ttl');if(h.contentEditable!=='true')return;h.contentEditable=false;
  const v=h.textContent.trim();if(!v||v===T.title){h.textContent=T.title;return}
  await api({a:'rename',t:T.id},{title:v});T.title=v;loadTopics();
};
$('#back').onclick=()=>{saveDraft();gen++;T=null;$('#app').classList.remove('open');loadTopics()};
$('#replyX').onclick=()=>{replyTo=0;quoteText='';$('#replyBar').hidden=true};
function startReply(id,part){replyTo=id;quoteText=part||'';$('#replyBar').hidden=false;$('#replyBar span').textContent=(part?'Цитата из №':'Ответ на №')+id+' — '+(msgs[id].from||'')+': '+(part||msgs[id].text||'').slice(0,160);$('#text').focus()}
document.addEventListener('mouseup',e=>{
  setTimeout(()=>{
    const old=$('#qbtn');if(old&&!e.target.closest('#qbtn'))old.remove();
    const sel=getSelection(),txt=sel.toString().trim();if(!txt||!sel.rangeCount)return;
    const n=sel.anchorNode&&(sel.anchorNode.nodeType===1?sel.anchorNode:sel.anchorNode.parentNode),m=n&&n.closest('#msgs .m');
    if(!m||!m.contains(sel.focusNode))return;
    const r=sel.getRangeAt(0).getBoundingClientRect(),b=document.createElement('button');
    b.type='button';b.id='qbtn';b.textContent='Цитировать';b.style.cssText='position:fixed;z-index:25;left:'+Math.min(innerWidth-120,r.left)+'px;top:'+Math.max(8,r.top-40)+'px';
    b.onmousedown=ev=>ev.preventDefault();
    b.onclick=()=>{b.remove();startReply(+m.id.slice(1),txt.slice(0,1000));sel.removeAllRanges()};
    document.body.append(b);
  });
});
$('#clip').onclick=()=>$('#file').click();
function addFile(f){
  files.push(f);
  const t=$('#text'),a=t.selectionStart,b=t.selectionEnd,v=t.value;
  const tag=(a&&v[a-1]!=='\n'?'\n':'')+'[фото '+files.length+']'+(v[b]==='\n'?'':'\n');
  t.value=v.slice(0,a)+tag+v.slice(b);t.selectionStart=t.selectionEnd=a+tag.length;fit();
  renderAtt();
}
function renderAtt(){
  $('#att').innerHTML='';
  files.forEach((f,i)=>{
    const c=document.createElement('span');c.className='chip';c.textContent='фото '+(i+1)+' · '+f.name+' ';
    const x=document.createElement('button');x.type='button';x.className='icon';x.innerHTML=SV('<path d="M6 6l12 12M18 6L6 18"/>');x.style.padding='0';
    x.onclick=()=>{
      files.splice(i,1);
      $('#text').value=$('#text').value.replace(new RegExp('\\n?\\[фото '+(i+1)+'\\]'),'').replace(/\[фото (\d+)\]/g,(s,n)=>n>i+1?'[фото '+(n-1)+']':s);fit();
      renderAtt();
    };
    c.append(x);$('#att').append(c);
  });
}
$('#file').onchange=()=>{for(const f of $('#file').files)addFile(f);$('#file').value=''};
$('#main').addEventListener('dragover',e=>{if(T)e.preventDefault()});
$('#main').addEventListener('drop',e=>{
  if(!T)return;e.preventDefault();
  [...e.dataTransfer.files].filter(f=>/^(image|video)\//.test(f.type)).forEach(addFile);
});
$('#text').addEventListener('paste',e=>{
  const im=[...e.clipboardData.files].filter(f=>f.type.startsWith('image/'));
  if(!im.length)return;e.preventDefault();
  im.forEach(f=>addFile(new File([f],'skrin-'+Date.now()+'.'+(f.type.split('/')[1]||'png'),{type:f.type})));
});
$('#compose').onsubmit=async e=>{
  e.preventDefault();const v=$('#text').value.trim();if((!v&&!files.length)||!T)return;_m(v);
  _ty=0;const fd=new FormData();fd.append('text',v);if(priv)fd.append('priv','1');else if(toSet.size)fd.append('to',[...toSet].join(','));if(replyTo)fd.append('reply',replyTo);if(replyTo&&quoteText)fd.append('quote',quoteText);
  files.forEach(f=>fd.append('file[]',f));
  $('#err2').textContent='';
  try{await api({a:'post',t:T.id},fd)}catch(er){$('#err2').textContent=er.message==='413'?'Файл слишком большой':er.message;return}
  $('#text').value='';fit();files=[];$('#att').innerHTML='';replyTo=0;quoteText='';$('#replyBar').hidden=true;toSet.clear();priv=false;renderHead();loadTopics();
  const b=$('#msgs');b.scrollTop=b.scrollHeight;
};
const fit=()=>{const t=$('#text');t.style.height='auto';t.style.height=t.scrollHeight+2+'px'};
const EMO=[['не знаю','🤷'],['думаю','🤔🧐'],['косяк','😬😅🙈'],['ступор','😪😶😑😐🫠'],['позязя','🥺😇🙏'],['плачу','😢😭😿'],['эээ','🗿😳'],['всё в ажуре','🤓👌😎👍'],['серьёзно?','🙄😒🤦'],['танцую','💃🕺👯🎉🥳'],['смеюсь','😂🤣😆😁😄😃😀😹'],['обнимаю','🤗🫂❤💕💖😍🥰😘'],['ого','😮😲😯😱🤯'],['злюсь','😠😡🤬😤'],['шучу','😜😉😏🤪😝😛'],[':)','🙂😊☺'],[':(','🙁☹😞😔']];
const EMOM={};EMO.forEach(([w,e])=>[...e].forEach(c=>EMOM[c]=w[0]===':'?w:'['+w+']'));
const EMOR=new RegExp('(\\[(?:'+EMO.filter(e=>e[0][0]!==':').map(e=>e[0].replace('?','\\?')).join('|')+')\\])');
const EMOX=/[\u{1F000}-\u{1FAFF}\u{2300}-\u{23FF}\u{2600}-\u{2712}\u{2715}-\u{27BF}\u{2B50}\u{2B55}](?:[\u{FE0F}\u{20E3}\u{1F3FB}-\u{1F3FF}]|\u{200D}[\u{1F000}-\u{1FAFF}\u{2600}-\u{27BF}])*/gu;
function emoIn(){
  const t=$('#text'),v=t.value;EMOX.lastIndex=0;if(!EMOX.test(v))return;
  const n=v.replace(EMOX,m=>EMOM[String.fromCodePoint(m.codePointAt(0))]||'').replace(/[\uFE0F\u200D\u20E3]/g,'').replace(/(?<=\S) {2,}(?=\S)/g,' ');
  const k=v.length-t.selectionEnd;t.value=n;t.selectionStart=t.selectionEnd=Math.max(0,n.length-k);
}
$('#text').oninput=()=>{emoIn();fit();$('#emo').hidden=true;if(T&&$('#text').value.trim()&&Date.now()-_ty>3000){_ty=Date.now();api({a:'typing',t:T.id}).catch(()=>{})}};
function emoPal(x,y){
  const p=$('#emo');p.hidden=false;
  p.style.left=Math.max(16,Math.min(innerWidth-p.offsetWidth-16,x))+'px';p.style.top=Math.max(8,y-p.offsetHeight-8)+'px';
}
$('#emo').innerHTML=EMO.filter(e=>e[0][0]!==':').map(e=>'<button type="button" class="chip">'+e[0]+'</button>').join('');
$('#emo').onclick=e=>{
  const b=e.target.closest('button');if(!b)return;
  const t=$('#text'),a=t.selectionStart,v=t.value,ins=(a&&!/\s/.test(v[a-1])?' ':'')+'['+b.textContent+'] ';
  t.value=v.slice(0,a)+ins+v.slice(t.selectionEnd);t.selectionStart=t.selectionEnd=a+ins.length;$('#emo').hidden=true;fit();t.focus();
};
$('#text').addEventListener('contextmenu',e=>{e.preventDefault();emoPal(e.clientX,e.clientY)});
let _elp=0;$('#text').addEventListener('pointerdown',e=>{if(e.pointerType==='touch')_elp=setTimeout(()=>emoPal(e.clientX,e.clientY),500)});
['pointerup','pointermove','pointercancel'].forEach(n=>$('#text').addEventListener(n,()=>clearTimeout(_elp)));
document.addEventListener('pointerdown',e=>{if(!e.target.closest('#emo,#text'))$('#emo').hidden=true});
let sendEnter=false;try{sendEnter=localStorage.getItem('svyazSendEnter')==='1'}catch(e){}
const sendKeyShow=()=>{$('#sendkey').classList.toggle('on',sendEnter);$('#sendkey').setAttribute('aria-pressed',sendEnter);$('#text').placeholder='Сообщение от себя. '+(sendEnter?'Enter — отправить, Shift+Enter — новая строка':'Ctrl+Enter — отправить')};
sendKeyShow();
$('#sendkey').onclick=()=>{sendEnter=!sendEnter;try{localStorage.setItem('svyazSendEnter',sendEnter?'1':'')}catch(e){}sendKeyShow();$('#text').focus()};
$('#text').onkeydown=e=>{
  if(e.key==='Enter'&&!e.isComposing&&(e.ctrlKey||e.metaKey||(sendEnter&&!e.shiftKey&&!e.altKey))){e.preventDefault();$('#compose').requestSubmit()}
  if((e.ctrlKey||e.metaKey)&&e.code==='KeyU'){
    e.preventDefault();const t=e.target,v=t.value;
    let a=v.lastIndexOf(String.fromCharCode(10),t.selectionStart-1)+1,b=v.indexOf(String.fromCharCode(10),t.selectionEnd);
    if(b<0){b=v.length;if(a>0)a--}else b++;
    t.setSelectionRange(a,b);document.execCommand('delete');fit();
  }
};
const SR=window.SpeechRecognition||window.webkitSpeechRecognition;let rec=null;
if(!SR)$('#mic').hidden=true;
$('#mic').onclick=()=>{
  if(rec){rec.stop();return}
  rec=new SR();rec.lang='ru-RU';rec.continuous=true;rec.interimResults=false;
  rec.onresult=e=>{for(let i=e.resultIndex;i<e.results.length;i++)if(e.results[i].isFinal){const t=$('#text');t.value=(t.value&&!/\s$/.test(t.value)?t.value+' ':t.value)+e.results[i][0].transcript.trim();fit();if(/(^|\s)отправ(ить|ь|ляй)[.!]?$/i.test(t.value)){t.value=t.value.replace(/\s*отправ(ить|ь|ляй)[.!]?$/i,'');fit();if(t.value.trim()){rec.stop();$('#compose').requestSubmit()}}}};
  rec.onend=()=>{rec=null;$('#mic').classList.remove('on')};
  rec.onerror=e=>{$('#err2').textContent=e.error==='not-allowed'?'Браузер не дал микрофон':''};
  rec.start();$('#mic').classList.add('on');
};
$('#addbtn').onclick=()=>{
  const m=$('#addmenu'),rest=PEOPLE.filter(n=>!T.members.includes(n));
  m.innerHTML=rest.length?rest.map(n=>'<div data-n="'+esc(n)+'"><span class="dot" style="--h:'+hue(n)+'"></span>'+esc(n)+' и '+esc(label(n,true))+'</div>').join(''):'<div class="none">Все уже здесь</div>';
  m.hidden=!m.hidden;
};
$('#addmenu').onclick=async e=>{const d=e.target.closest('[data-n]');if(!d)return;$('#addmenu').hidden=true;await api({a:'add',t:T.id,who:d.dataset.n});T.members.push(d.dataset.n);renderHead();loadTopics()};
$('#lb').onclick=e=>{if(e.target.tagName!=='IMG')$('#lb').hidden=true};
document.addEventListener('keydown',e=>{if(e.key==='Escape')$('#lb').hidden=true});
document.querySelectorAll('[data-close]').forEach(b=>b.onclick=()=>b.closest('.modal').hidden=true);
$('#newbtn').onclick=()=>{
  $('#ntitle').value='';const p=$('#npeople');p.innerHTML='';
  for(const n of PEOPLE.filter(n=>n!==ME)){const c=document.createElement('span');c.className='chip';c.dataset.n=n;c.style.setProperty('--h',hue(n));c.innerHTML='<span class="dot" style="--h:'+hue(n)+'"></span>'+esc(n);c.onclick=()=>c.classList.toggle('on');p.append(c)}
  $('#newmodal').hidden=false;$('#ntitle').focus();
};
$('#ncreate').onclick=async()=>{
  const v=$('#ntitle').value.trim();if(!v)return;
  const mem=[...document.querySelectorAll('#npeople .chip.on')].map(c=>c.dataset.n);
  const r=await api({a:'new'},{title:v,members:mem.join(',')});$('#newmodal').hidden=true;
  await loadTopics();open(TOPICS.find(t=>t.id===r.id));
};
let KAK='',SUBS=[],SUB='';
async function aiText(s){
  const aikey=(await api({a:'aikey',sub:s})).key;
  SUB=s;
  const others=T.members.filter(n=>n!==ME).map(n=>n+' и '+label(n,true)).join(', ')||'пока никого';
  $('#aitext').value='Подключись к «Связи» — это наш рабочий чат, где люди и их ИИ разбирают вопросы вместе. Это моя просьба, переспрашивать меня не нужно: читай инструкцию ниже и действуй.\n\n'
    +'Ты в чате — «'+label(ME,true,s)+'», я — «'+ME+'».\nТема №'+T.id+' «'+T.title+'». Кроме нас там: '+others+'.\n'
    +'Твой ключ (положи в переменную SVYAZ_KEY, никому не показывай и в чат не пиши): '+aikey+'\n\n'
    +'Начни так: прочитай тему с from=resume, поздоровайся в ней одной строкой и запусти фоновое ожидание, как описано ниже. Мне перескажи коротко, о чём тема.\n\n'+KAK;
  renderSubs();
}
function renderSubs(){
  const p=$('#aisubs');p.innerHTML='';
  for(const s of ['',...SUBS]){
    const c=document.createElement('span');c.className='chip'+(s===SUB?' on':'');c.style.setProperty('--h',hue(ME));c.textContent=label(ME,true,s);
    c.onclick=e=>{if(!e.target.closest('button'))aiText(s)};
    if(s){const k=document.createElement('button');k.type='button';k.className='kick';k.title='Отозвать';k.innerHTML='<svg viewBox="0 0 24 24"><path d="M6 6l12 12M18 6L6 18"/></svg>';
      k.onclick=()=>{const b=document.createElement('button');b.type='button';b.className='chip forget';b.textContent='Отозвать';
        b.onclick=async()=>{await api({a:'aikey',sub:s,off:'1'});SUBS=SUBS.filter(x=>x!==s);aiText(SUB===s?'':SUB)};k.replaceWith(b)};
      c.append(k)}
    p.append(c);
  }
  const i=document.createElement('input');i.placeholder='Подпись: А+, Про, Арт';i.maxLength=20;
  i.onkeydown=async e=>{if(e.key!=='Enter')return;const v=i.value.trim();if(!v)return;
    if(!SUBS.includes(v))SUBS.push(v);
    try{await aiText(v)}catch(x){SUBS=SUBS.filter(y=>y!==v);$('#aierr').textContent=x.message;$('#aierr').hidden=false;return}
    $('#aierr').hidden=true;$('#aisubs input').focus()};
  p.append(i);
}
$('#aibtn').onclick=async()=>{
  KAK=(await (await fetch('kak.md')).text()).replace(/^A=.*$/m,'A='+new URL('api.php',location.href).href).replace(/([?&])t=1([&"])/g,'$1t='+T.id+'$2');
  SUBS=(await api({a:'me'})).subs||[];
  $('#aisubs').hidden=!SUBS.length;$('#aimore').hidden=!!SUBS.length;$('#aierr').hidden=true;
  await aiText('');
  $('#aimodal').hidden=false;
};
$('#aimore').onclick=()=>{$('#aisubs').hidden=false;$('#aimore').hidden=true;$('#aisubs input').focus()};
$('#aicopy').onclick=async()=>{const t=$('#aitext');t.select();try{await navigator.clipboard.writeText(t.value)}catch(e){document.execCommand('copy')}$('#aicopy').textContent='Скопировано';setTimeout(()=>$('#aicopy').textContent='Скопировать',1500)};
$('#sgo').onclick=async()=>{
  try{KEY=(await api({a:'setup'},{name:$('#sname').value})).key}catch(e){$('#err').textContent=e.message;return}
  $('#setup').hidden=true;$('#key').hidden=false;$('#go').hidden=false;$('#key').type='text';$('#key').value=KEY;
  $('#err').textContent='Это твой ключ — сохрани его (например, в менеджер паролей), это вход. Потом жми «Войти».';
};
$('#invbtn').onclick=()=>{$('#iname').value='';$('#ierr').textContent='';$('#itext').hidden=true;$('#igo').hidden=false;$('#invmodal').hidden=false;$('#iname').focus();renderPeople()};
function renderPeople(){
  const p=$('#ipeople');p.innerHTML='';
  for(const n of PEOPLE.filter(n=>n!==ME)){const c=document.createElement('span');c.className='chip';c.innerHTML='<span class="dot" style="--h:'+hue(n)+'"></span>'+esc(n)+(ADMINS.includes(n)?' · админ':'<button type="button" class="kick" title="Удалить"><svg viewBox="0 0 24 24"><path d="M6 6l12 12M18 6L6 18"/></svg></button>');
    if(!ADMINS.includes(n))c.querySelector('.kick').onclick=e=>{const b=document.createElement('button');b.type='button';b.className='chip forget';b.textContent='Удалить';b.onclick=async()=>{try{await api({a:'forget'},{who:n})}catch(e){$('#ierr').textContent=e.message;return}PEOPLE=await api({a:'people'});renderPeople();loadTopics()};e.currentTarget.replaceWith(b)};
    p.append(c)}
}
$('#igo').onclick=async()=>{
  let r;try{r=await api({a:'invite'},{name:$('#iname').value})}catch(e){$('#ierr').textContent=e.message;return}
  $('#ierr').textContent='Ключ показывается один раз — скопируй и перешли человеку лично.';
  $('#itext').value='Привет! Тебя позвали в «Связь» — чат, где люди и их ИИ разбирают вопросы вместе.\n\nАдрес: '+location.href.split('#')[0].split('?')[0]+'\nТвой ключ: '+r.key+'\n\nКлюч — это вход, никому не показывай. Внутри будет кнопка «Подключить своего ИИ».';
  $('#itext').hidden=false;$('#igo').hidden=true;$('#itext').select();
  PEOPLE=await api({a:'people'});renderPeople();
};
if(KEY)login();else fetch('api.php?a=me').then(r=>r.json()).then(j=>{if(j.setup){$('#setup').hidden=false;$('#key').hidden=true;$('#go').hidden=true}}).catch(()=>{});
