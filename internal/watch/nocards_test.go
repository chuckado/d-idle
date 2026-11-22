package watch

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"d2lod-bot/internal/goals"
	"d2lod-bot/internal/unlock"
)

// The identify-reveal card feature is gone (user report 2026-11-19: no "unidentified item" card anywhere). 50
// unidentified drops, on two classes, must produce no card route, no card toast or event line and no card payload; the
// page and the presets carry no Cards tab; unlocking never names cards.
func TestNoUnidentifiedCardFromDrops(t *testing.T) {
	for _, class := range []string{"Sorceress", "Paladin"} {
		t.Run(class, func(t *testing.T) {
			l, err := unlock.Shipped()
			if err != nil {
				t.Fatal(err)
			}
			eng, err := unlock.Open(t.TempDir(), strings.ToLower(class), l, nil)
			if err != nil {
				t.Fatal(err)
			}
			r := newFunRigWith(t, class, 1, func(_ *funRig, fc *FunConfig) { fc.Unlock = eng })
			for i := 1; i <= 50; i++ {
				q := []string{"magic", "rare", "unique", "set"}[i%4]
				r.feed(goals.Event{Type: goals.EvDrop, Seq: uint64(i), Item: &goals.Item{Name: "Short Sword", Quality: q}})
			}
			if eng.Allowed("idreveal") || eng.Allowed("idreveal_later") {
				t.Error("the ladder still has a card row")
			}
			for _, route := range []string{"/api/reveal", "/api/reveal/flip"} {
				for _, base := range []string{r.view, r.answer} {
					if resp, _ := r.do("GET", base, route, funToken, ""); resp.StatusCode == 200 {
						t.Errorf("%s answers %d, want no card route", route, resp.StatusCode)
					}
				}
			}
			for _, route := range []string{"/api/unlocks", "/api/away", "/api/chronicle", "/api/needs"} {
				_, body := r.do("GET", r.view, route, funToken, "")
				for _, bad := range []string{"Unidentified item", "face-down", "Flip the card", "idreveal"} {
					if strings.Contains(string(body), bad) {
						t.Errorf("%s mentions %q", route, bad)
					}
				}
			}
		})
	}
}

func TestShellHasNoCardsTab(t *testing.T) {
	ents, err := webFS.ReadDir("web/layouts")
	if err != nil || len(ents) == 0 {
		t.Fatal(err)
	}
	for _, e := range ents {
		if e.Name() == "classic.json" || e.Name() == "mockup.json" || e.Name() == "live.json" || e.Name() == "command.json" {
			continue // frozen (hashed) files keep the id; shell.json reserves it, so it mounts nothing
		}
		b, _ := webFS.ReadFile("web/layouts/" + e.Name())
		if strings.Contains(string(b), `"id": "cards"`) {
			t.Errorf("preset %s has a cards panel", e.Name())
		}
	}
	h, _ := webFS.ReadFile("web/watch.html")
	for _, bad := range []string{`id="tab-cards"`, "revealsheet", "cardspane", "Unidentified item"} {
		if strings.Contains(string(h), bad) {
			t.Errorf("watch.html still has %q", bad)
		}
	}
	for _, f := range []string{"web/unlockui.js", "web/inventory-actions.js"} {
		b, _ := webFS.ReadFile(f)
		if strings.Contains(string(b), "/api/reveal/flip") || strings.Contains(string(b), "'Unidentified item'") {
			t.Errorf("%s still builds the card", f)
		}
	}
}

// unlockui.js in a stub DOM (node, skipped without it): even a server that still answers a card list and an open
// "idreveal" row makes the script touch no card element and open no sheet.
func TestUnlockUIBuildsNoCards(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("no node")
	}
	js, _ := webFS.ReadFile("web/unlockui.js")
	harness := `
var touched={};
function mk(t){var e={tag:t,children:[],firstChild:null,className:'',textContent:'',
 appendChild:function(c){this.children.push(c);this.firstChild=this.children[0];return c;},
 removeChild:function(c){this.children.splice(this.children.indexOf(c),1);this.firstChild=this.children[0]||null;},
 addEventListener:function(){},setAttribute:function(){},focus:function(){},
 classList:{toggle:function(){},add:function(){}},querySelector:function(){return null;},checked:true,hidden:true};
 return e;}
global.document={getElementById:function(id){touched[id]=1;var e=mk('div');e.id=id;return e;},
 createElement:mk,addEventListener:function(){},querySelector:function(){return null;}};
global.window={};global.location={protocol:'http:',hostname:'h'};
var fetched=[];
function ok(o){return Promise.resolve({ok:true,json:function(){return Promise.resolve(o);}});}
global.fetch=function(u){fetched.push(u);
 if(u==='/api/unlocks')return ok({open:[{id:'idreveal'}],locked:[]});
 return ok({});};
` + string(js) + `
(async function(){
 await new Promise(function(r){setTimeout(r,60);});
 var bad=Object.keys(touched).filter(function(k){return /^(tab-cards|cardspane|cardlist|revealsheet|revealsound|rv-)/.test(k);});
 var card=fetched.filter(function(u){return /reveal/.test(u);});
 console.log(JSON.stringify({bad:bad,card:card}));
 process.exit(0);
})();`
	f := filepath.Join(t.TempDir(), "h.js")
	os.WriteFile(f, []byte(harness), 0o600)
	b, err := exec.Command(node, f).CombinedOutput()
	if err != nil {
		t.Fatalf("%v: %s", err, b)
	}
	if !strings.Contains(string(b), `{"bad":[],"card":[]}`) {
		t.Errorf("the page script still builds cards: %s", b)
	}
}
