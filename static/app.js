const el = document.getElementById('app');
let cur, state, page = 'dashboard',
  selected = null,
  search = '';
const esc = s => String(s ?? '').replace(/[&<>"']/g, c => ({
  '&': '&amp;',
  '<': '&lt;',
  '>': '&gt;',
  '"': '&quot;',
  "'": '&#39;'
} [c]));
const all = () => cur.stages.flatMap(s => s.modules);
const prog = id => state.progress[id] || {
  status: 'not_started',
  notes: ''
};
const pct = s => Math.round(100 * s.modules.filter(m => prog(m.id).status === 'passed').length / s.modules
  .length);
const tot = () => all().length;
const passed = () => all().filter(m => prog(m.id).status === 'passed').length;
const mins = () => state.sessions.reduce((t, x) => t + x.minutes, 0);
const statusLabels = {
  not_started: 'Not started',
  in_progress: 'In progress',
  ready_for_assessment: 'Ready for assessment',
  passed: 'Passed',
  needs_revision: 'Needs revision'
};
const localDate = () => {
  const d = new Date();
  return `${d.getFullYear()}-${String(d.getMonth()+1).padStart(2,'0')}-${String(d.getDate()).padStart(2,'0')}`
};
const badge = (status) => `<span class="tag ${esc(status)}">${statusLabels[status]||'Not started'}</span>`;
async function api(path, options) {
  const r = await fetch(path, options);
  const j = await r.json();
  if (!r.ok) throw Error(j.error || `HTTP ${r.status}`);
  return j
}

function nav(p, id) {
  page = p;
  selected = id || null;
  search = '';
  render();
  window.scrollTo({
    top: 0,
    behavior: 'instant'
  })
}

function sidebar() {
  return `<aside class="sidebar"><div class="brand"><span class="brandlogo">✈</span><span>ROTORCRAFT<br>ACADEMY</span></div><div class="smallcaps">YOUR WORKSPACE</div>${[['dashboard','◫','Overview'],['curriculum','▤','Curriculum'],['sessions','◷','Study log'],['resources','▣','Resources']].map(([p,i,n])=>`<button class="navbtn ${page===p?'active':''}" data-nav="${p}">${i} &nbsp;${n}</button>`).join('')}<div class="footer">SINGLE USER • LOCAL FIRST<br>Helicopter engineering · v1</div></aside>`
}

function metrics() {
  return `<div class="cards"><div class="card"><div class="statlabel">Curriculum completion</div><div class="value">${Math.round(passed()/tot()*100)}<small>%</small></div><div class="bar"><span style="width:${passed()/tot()*100}%"></span></div></div><div class="card"><div class="statlabel">Modules passed</div><div class="value">${passed()} <small>/ ${tot()}</small></div></div><div class="card"><div class="statlabel">Study time logged</div><div class="value">${(mins()/60).toFixed(1)} <small>hours</small></div></div><div class="card"><div class="statlabel">Total planned study</div><div class="value">${cur.stages.reduce((n,s)=>n+s.hours,0).toLocaleString()} <small>hours</small></div></div></div>`
}

function dashboard() {
  let next = all().find(m => prog(m.id).status === 'in_progress') || all().find(m => prog(m.id).status !==
    'passed');
  return `<div class="smallcaps">FLIGHT ACADEMY / DASHBOARD</div><h1>Engineering your way to flight.</h1><p class="intro">A staged helicopter engineering curriculum outline, from fundamentals to a concept-level rotorcraft design. Track real competence—not just reading time.</p>${metrics()}<div class="split"><section class="panel"><div class="stagehead"><h2>Learning pathway</h2><button class="btn secondary" data-nav="curriculum">All courses →</button></div>${cur.stages.map((s,i)=>`<div class="stage"><div class="stagehead"><strong>${String(i+1).padStart(2,'0')} &nbsp;${esc(s.name)}</strong><small>${pct(s)}%</small></div><div class="stagemeta">${s.modules.length} modules · ${s.hours} planned hours</div><div class="bar"><span style="width:${pct(s)}%"></span></div></div>`).join('')}</section><div><section class="hero"><div class="smallcaps">CONTINUE LEARNING</div><h2>${esc(next?.title||'Curriculum complete!')}</h2><p>${next?esc(next.criteria):'All modules have been marked passed.'}</p>${next?`<button class="btn" data-module="${next.id}">Open lesson →</button>`:''}</section><section class="panel"><h3>Recently logged</h3>${state.sessions.slice(-4).reverse().map(x=>`<div class="session"><strong>${esc(x.moduleId)}</strong><div class="note">${esc(x.date)} · ${x.minutes} min</div><p>${esc(x.note)}</p></div>`).join('')||'<p>No sessions yet. Start with the first module and record your study.</p>'}<button class="btn secondary" data-nav="sessions">Log study time</button></section></div></div>`
}

function curriculum() {
  return `<div class="smallcaps">STRUCTURED LEARNING</div><h1>Curriculum</h1><p class="intro">Eight stages, ${tot()} modules, prerequisites, textbooks, engineering calculations, and assessment requirements.</p><input class="search" id="search" aria-label="Search curriculum" placeholder="Search topics, skills, resources…" value="${esc(search)}"><div id="coursegrid" class="coursegrid">${courseCards()}</div>`
}

function courseCards() {
  const q = search.toLowerCase();
  return cur.stages.map((s, i) => ({
    s,
    i
  })).filter(({
    s
  }) => (s.name + ' ' + s.description + ' ' + s.modules.map(m => m.title + ' ' + m.resources + ' ' + m
    .criteria).join(' ')).toLowerCase().includes(q)).map(({
      s,
      i
    }) =>
    `<button class="panel coursecard" data-stage="${s.id}"><div class="smallcaps">STAGE ${String(i+1).padStart(2,'0')}</div><h3>${esc(s.name)}</h3><p>${esc(s.description)}</p><div class="stagehead"><small>${s.modules.length} modules · ${s.hours} h</small><strong style="color:var(--accent)">${pct(s)}%</strong></div><div class="bar"><span style="width:${pct(s)}%"></span></div></button>`
    ).join('') || '<p>No matching courses.</p>'
}

function stageView() {
  const s = cur.stages.find(x => x.id === selected);
  return `<button class="back" data-nav="curriculum">← All courses</button><div class="smallcaps">${esc(s.id)} / LEARNING PATH</div><h1>${esc(s.name)}</h1><p class="intro">${esc(s.description)}</p><div class="cards"><div class="card"><div class="statlabel">Modules</div><div class="value">${s.modules.length}</div></div><div class="card"><div class="statlabel">Study workload</div><div class="value">${s.hours}<small> h</small></div></div><div class="card"><div class="statlabel">Completed</div><div class="value">${pct(s)}<small>%</small></div></div></div><div class="modulelist">${s.modules.map((m,i)=>`<button class="module" data-module="${m.id}"><div><span class="statuscircle ${prog(m.id).status}"></span><strong style="display:inline">${i+1}. ${esc(m.title)}</strong><div class="stagemeta" style="margin:10px 0 0 20px">${m.hours} study hours · ${esc(m.id)} ${m.prerequisites.length?' · prerequisite: '+m.prerequisites.join(', '):''}</div></div>${badge(prog(m.id).status)}</button>`).join('')}</div>`
}

function lesson() {
  const m = all().find(x => x.id === selected),
    pr = prog(m.id),
    stage = cur.stages.find(s => s.modules.some(x => x.id === m.id));
  return `<div class="details"><button class="back" data-stage="${esc(stage.id)}">← Back to stage</button><div class="smallcaps">MODULE ${m.id} · ${m.hours} HOURS</div><h1>${esc(m.title)}</h1><div class="topactions">${badge(pr.status)} ${m.prerequisites.length?`<span class="muted">Prerequisite: ${m.prerequisites.map(id=>`${esc(id)} (${prog(id).status==='passed'?'passed':'not passed'})`).join(', ')}</span>`:''}</div><section class="section"><h2>Study resources</h2><p>${esc(m.resources)}</p><p class="note">${m.resourcesVerification==='verified'?'Resource verification is recorded in the curriculum.':'Requires verification: edition, chapter/section locations and source links have not been checked. Treat these references as topic guidance.'}</p></section><section class="section"><h2>Passing criteria</h2><p>${esc(m.criteria)}</p><h3>Assessment standard</h3><p>${esc(m.assessment)}</p><p class="note">Manual grading: passing requires a score ≥80%, evidence references, confirmation that mandatory practical work and calculations were reviewed, and passed prerequisites. This records a human assessment; the app does not validate engineering designs.</p></section><section class="section"><h2>Update assessment</h2><form id="progressForm"><div class="formrow"><label class="field">Status<select name="status">${Object.keys(statusLabels).map(x=>`<option value="${x}" ${pr.status===x?'selected':''}>${statusLabels[x]}</option>`).join('')}</select></label><label class="field">Assessment score (0–100)<input type="number" min="0" max="100" step="0.1" name="score" value="${esc(pr.score??'')}" placeholder="Score"></label></div><label class="field">Study notes<textarea maxlength="5000" name="notes" placeholder="Worked problems, project link, issues to revisit…">${esc(pr.notes)}</textarea></label><label class="field">Submitted work / evidence references<textarea name="evidence" maxlength="5000" placeholder="Report path, CAD revision, simulation commit or external portfolio URL…">${esc(pr.evidence)}</textarea></label><label class="check"><input type="checkbox" name="practicalConfirmed" ${pr.practicalConfirmed?'checked':''}> Mandatory calculations and practical work reviewed; critical errors resolved.</label>${pr.completedAt?`<p class="note">Passed on ${esc(new Date(pr.completedAt).toLocaleString())}</p>`:''}<button class="btn" type="submit">Save progress</button><span id="progressMessage" role="status" aria-live="polite" style="margin-left:12px"></span></form></section><section class="section"><h2>Record study session</h2><form id="sessionForm"><div class="formrow"><label class="field">Minutes studied<input type="number" name="minutes" min="1" max="1440" required value="45"></label><label class="field">Date<input type="date" name="date" value="${localDate()}"></label></div><label class="field">What did you work on?<textarea maxlength="2000" name="note" placeholder="e.g. derived the induced power equation"></textarea></label><button class="btn secondary" type="submit">Add study session</button><span id="sessionMessage" role="status" aria-live="polite" style="margin-left:12px"></span></form></section></div>`
}

function sessions() {
  return `<div class="smallcaps">LEARNING ACTIVITY</div><h1>Study log</h1><p class="intro">Every session is saved locally. Progress and study hours survive application restarts.</p>${metrics()}<div class="panel"><div class="stagehead"><h2>Study history</h2><button id="export" class="btn secondary">Export JSON</button></div>${state.sessions.slice().reverse().map(x=>`<div class="session"><div class="stagehead"><strong>${esc(all().find(m=>m.id===x.moduleId)?.title||x.moduleId)}</strong><span class="tag">${x.minutes} minutes</span></div><div class="muted">${esc(x.date)}</div><p>${esc(x.note)}</p></div>`).join('')||'<p>No sessions logged yet. Open any module to record study time.</p>'}</div>`
}

function resources() {
  return `<div class="smallcaps">REFERENCE MATERIALS</div><h1>Engineering library</h1><p class="intro">References are organized within each module. Some books require purchase or library access.</p><div class="panel"><h2>Free starting points</h2><p><a href="https://www.faa.gov/regulations_policies/handbooks_manuals/aviation/helicopter_flying_handbook" target="_blank" rel="noopener">FAA Helicopter Flying Handbook ↗</a></p><p><a href="https://ocw.mit.edu/courses/18-03sc-differential-equations-fall-2011/" target="_blank" rel="noopener">MIT Differential Equations ↗</a></p><p><a href="https://ocw.mit.edu/courses/18-06-linear-algebra-spring-2010/" target="_blank" rel="noopener">MIT Linear Algebra ↗</a></p><p><a href="https://www.nasa.gov/reference/systems-engineering-handbook/" target="_blank" rel="noopener">NASA Systems Engineering Handbook ↗</a></p><p><a href="https://my.solidworks.com/training" target="_blank" rel="noopener">Official SOLIDWORKS Training ↗</a></p><p class="note">Curriculum references are learning guidance, not an approved aircraft design process.</p></div>`
}

function render() {
  el.innerHTML =
    `<div class="layout">${sidebar()}<main class="main">${({dashboard,curriculum,stage:stageView,lesson,sessions,resources})[page]()}</main></div>`;
  el.querySelectorAll('[data-nav]').forEach(b => b.onclick = () => nav(b.dataset.nav));
  el.querySelectorAll('[data-stage]').forEach(b => b.onclick = () => nav('stage', b.dataset.stage));
  el.querySelectorAll('[data-module]').forEach(b => b.onclick = () => nav('lesson', b.dataset.module));
  const sr = el.querySelector('#search');
  if (sr) sr.oninput = e => {
    search = e.target.value;
    el.querySelector('#coursegrid').innerHTML = courseCards();
    el.querySelectorAll('[data-stage]').forEach(b => b.onclick = () => nav('stage', b.dataset.stage))
  };
  const pf = el.querySelector('#progressForm');
  if (pf) pf.onsubmit = async e => {
    e.preventDefault();
    const f = new FormData(pf),
      msg = el.querySelector('#progressMessage'),
      moduleId = selected,
      submit = pf.querySelector('[type=submit]');
    submit.disabled = true;
    try {
      const score = f.get('score');
      state.progress[moduleId] = await api('/api/progress/' + encodeURIComponent(moduleId), {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          status: f.get('status'),
          score: score === '' ? null : Number(score),
          notes: f.get('notes'),
          evidence: f.get('evidence'),
          practicalConfirmed: f.get('practicalConfirmed') === 'on'
        })
      });
      if (page !== 'lesson' || selected !== moduleId) return;
      render();
      el.querySelector('#progressMessage').textContent = 'Saved';
      el.querySelector('#progressMessage').className = 'success'
    } catch (ex) {
      msg.textContent = ex.message;
      msg.className = 'error'
    } finally {
      submit.disabled = false
    }
  };
  const sf = el.querySelector('#sessionForm');
  if (sf) sf.onsubmit = async e => {
    e.preventDefault();
    const f = new FormData(sf),
      msg = el.querySelector('#sessionMessage'),
      moduleId = selected,
      submit = sf.querySelector('[type=submit]');
    submit.disabled = true;
    try {
      const x = await api('/api/sessions', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          moduleId,
          minutes: Number(f.get('minutes')),
          date: f.get('date'),
          note: f.get('note')
        })
      });
      state.sessions.push(x);
      if (page === 'lesson' && selected === moduleId) {
        render();
        const saved = el.querySelector('#sessionMessage');
        saved.textContent = 'Session saved';
        saved.className = 'success'
      }
    } catch (ex) {
      msg.textContent = ex.message;
      msg.className = 'error'
    } finally {
      submit.disabled = false
    }
  };
  const exp = el.querySelector('#export');
  if (exp) exp.onclick = () => window.location.assign('/api/export')
};
Promise.all([api('/api/curriculum'), api('/api/state')]).then(([c, s]) => {
  cur = c;
  state = s;
  render()
}).catch(e => {
  el.innerHTML = '<div class="loading error">Could not load portal: ' + esc(e.message) + '</div>'
});
