package watch

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// skills.js in a stub DOM (node, skipped without it): every tile of known kind carries exactly one badge with text, the
// chip counts add up to the total, a filter hides the other kinds, and the choice is stored.
func TestSkillsJSKindBadgesAndFilter(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("no node")
	}
	js, _ := webFS.ReadFile("web/skills.js")
	view := `{"available":true,"class":"Paladin","level":30,"unspent":0,"spent":3,"allocate":false,"trees":[{"tab":1,"spent":3,"w":300,"h":504,"nodes":[` +
		`{"id":1,"name":"Sacrifice","x":0,"y":0,"w":64,"h":64,"state":"learned","base":1,"maxLevel":20,"level":1,"kind":"active"},` +
		`{"id":2,"name":"Zeal","x":0,"y":0,"w":64,"h":64,"state":"learned","base":1,"maxLevel":20,"level":1,"kind":"active"},` +
		`{"id":3,"name":"Might","x":0,"y":0,"w":64,"h":64,"state":"learned","base":1,"maxLevel":20,"level":1,"kind":"aura","aura":true},` +
		`{"id":4,"name":"Warmth","x":0,"y":0,"w":64,"h":64,"state":"locked","base":0,"maxLevel":20,"level":0,"kind":"passive","passive":true},` +
		`{"id":5,"name":"Mystery","x":0,"y":0,"w":64,"h":64,"state":"locked","base":0,"maxLevel":20,"level":0}]}]}`
	harness := `
var store={}, timers=[];
function mk(t){var e={tag:t,children:[],attrs:{},hidden:false,className:'',_text:'',listeners:{},style:{setProperty:function(){}},
 get textContent(){return this._text;}, set textContent(v){this._text=v;this.children=[];},
 get firstChild(){return this.children[0]||null;},
 appendChild:function(c){this.children.push(c);return c;},removeChild:function(c){this.children=this.children.filter(function(x){return x!==c;});},
 remove:function(){},setAttribute:function(k,v){this.attrs[k]=String(v);},getAttribute:function(k){return k in this.attrs?this.attrs[k]:null;},
 removeAttribute:function(k){delete this.attrs[k];},addEventListener:function(n,f){this.listeners[n]=f;}};return e;}
var ids={msg:mk('p'),summeta:mk('span'),trees:mk('div'),kindbar:mk('div')};
global.document={getElementById:function(i){return ids[i]||null;},createElement:mk,createElementNS:function(n,t){return mk(t);},addEventListener:function(){},
 querySelectorAll:function(){return [];},body:mk('body'),documentElement:{clientWidth:400}};
global.window={innerHeight:800,confirm:function(){return false;}};
global.Image=function(){return mk('img');};
global.localStorage={getItem:function(k){return k in store?store[k]:null;},setItem:function(k,v){store[k]=v;}};
global.location={protocol:'http:',hostname:'h'};
global.setInterval=function(){};
global.fetch=function(){return Promise.resolve({ok:true,status:200,text:function(){return Promise.resolve(VIEW);}});};
var VIEW=` + "`" + view + "`" + `;
` + string(js) + `
function all(e,o){o=o||[];o.push(e);e.children.forEach(function(c){all(c,o);});return o;}
function cls(e,c){return (' '+e.className+' ').indexOf(' '+c+' ')>=0;}
function snap(){
 var tiles=all(ids.trees).filter(function(e){return cls(e,'sk-tile');});
 return {tiles:tiles.map(function(t){
   var b=all(t).filter(function(e){return cls(e,'sk-kind');});
   return {name:t.attrs['data-id'],hidden:t.hidden,badges:b.length,text:b.length?all(b[0]).filter(function(e){return cls(e,'sk-kt');}).map(function(e){return e.textContent;}).join(''):'',label:t.attrs['aria-label']};}),
  chips:ids.kindbar.children.map(function(c){return {k:c.attrs['data-kind'],txt:c.children.map(function(x){return x.textContent;}).join(' '),pressed:c.attrs['aria-pressed']};}),
  stored:store['sk.kind']||null};
}
function tick(){return new Promise(function(r){setTimeout(r,20);});}
(async function(){
 await tick(); await tick();
 var out={all:snap()};
 ids.kindbar.children[3].listeners.click(); out.passive=snap();
 ids.kindbar.children[1].listeners.click(); out.active=snap();
 console.log(JSON.stringify(out));
})();`
	f := filepath.Join(t.TempDir(), "h.js")
	os.WriteFile(f, []byte(harness), 0o600)
	b, err := exec.Command(node, f).CombinedOutput()
	if err != nil {
		t.Fatalf("%v: %s", err, b)
	}
	type tile struct {
		Name   string
		Hidden bool
		Badges int
		Text   string
		Label  string
	}
	type chip struct{ K, Txt, Pressed string }
	type snapT struct {
		Tiles  []tile
		Chips  []chip
		Stored string
	}
	var out map[string]snapT
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("%v: %s", err, b)
	}
	a := out["all"]
	want := map[string]string{"1": "Active", "2": "Active", "3": "Aura", "4": "Passive", "5": ""}
	for _, tl := range a.Tiles {
		w := want[tl.Name]
		if w == "" && tl.Badges != 0 {
			t.Errorf("unknown kind must carry no badge: %+v", tl)
		}
		if w != "" && (tl.Badges != 1 || tl.Text != w) {
			t.Errorf("tile %s: want one %q badge: %+v", tl.Name, w, tl)
		}
		if tl.Hidden {
			t.Errorf("All hides nothing: %+v", tl)
		}
	}
	wantChips := []string{"All 5", "Active 2", "Aura 1", "Passive 1"}
	if len(a.Chips) != 4 {
		t.Fatalf("chips: %+v", a.Chips)
	}
	for i, c := range a.Chips {
		if c.Txt != wantChips[i] {
			t.Errorf("chip %d: %q, want %q", i, c.Txt, wantChips[i])
		}
	}
	if a.Chips[0].Pressed != "true" {
		t.Errorf("All starts pressed: %+v", a.Chips)
	}
	hid := func(s snapT) (n int) {
		for _, tl := range s.Tiles {
			if tl.Hidden {
				n++
			}
		}
		return
	}
	// 2+1+1 known kinds; the unknown tile is hidden by every kind filter
	if p := out["passive"]; hid(p) != 4 || p.Stored != "passive" || p.Chips[3].Pressed != "true" {
		t.Errorf("passive filter: %+v", p)
	}
	if p := out["active"]; hid(p) != 3 || p.Stored != "active" {
		t.Errorf("active filter: %+v", p)
	}
}
