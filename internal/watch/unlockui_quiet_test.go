package watch

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// unlockui.js in a stub DOM (node, skipped without it): 50 unflipped face-down cards, over several polls and a
// "reload", must never open the reveal sheet by themselves. They show as one count on the Cards tab; only a click on
// a card's button opens the sheet. Two item kinds (magic, rare) are in the list.
func TestUnlockUIUnidentifiedCardsNeverPopUp(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("no node")
	}
	js, _ := webFS.ReadFile("web/unlockui.js")
	harness := `
var store={},sheetOpens=0;
function mk(t){var e={tag:t,children:[],firstChild:null,className:'',textContent:'',listeners:{},_h:false,
 appendChild:function(c){this.children.push(c);this.firstChild=this.children[0];return c;},
 removeChild:function(c){this.children.splice(this.children.indexOf(c),1);this.firstChild=this.children[0]||null;},
 addEventListener:function(n,f){this.listeners[n]=f;},setAttribute:function(){},focus:function(){},
 classList:{toggle:function(){},add:function(){}},querySelector:function(){return null;},checked:true};
 Object.defineProperty(e,'hidden',{get:function(){return this._h;},set:function(v){this._h=v;if(this.id==='revealsheet'&&v===false)sheetOpens++;}});
 return e;}
global.document={getElementById:function(id){if(!store[id]){store[id]=mk('div');store[id].id=id;store[id]._h=true;}return store[id];},
 createElement:mk,addEventListener:function(){},querySelector:function(){return null;}};
global.window={};global.location={protocol:'http:',hostname:'h'};
var cards=[];for(var i=1;i<=50;i++)cards.push({id:'c'+i,n:i,revealed:false,quality:i%2?'magic':'rare'});
function ok(o){return Promise.resolve({ok:true,json:function(){return Promise.resolve(o);}});}
global.fetch=function(u){
 if(u==='/api/unlocks')return ok({open:[{id:'idreveal'}],locked:[]});
 if(u==='/api/reveal')return ok({face:'Unidentified item',faceText:'x',cards:cards});
 return ok({});};
` + string(js) + `
function tick(){return new Promise(function(r){setTimeout(r,30);});}
(async function(){
 await tick();await tick();
 console.log(JSON.stringify({opens:sheetOpens,tab:store['tab-cards'].textContent}));
 process.exit(0);
})();`
	f := filepath.Join(t.TempDir(), "h.js")
	os.WriteFile(f, []byte(harness), 0o600)
	b, err := exec.Command(node, f).CombinedOutput()
	if err != nil {
		t.Fatalf("%v: %s", err, b)
	}
	s := string(b)
	if !strings.Contains(s, `"opens":0`) || !strings.Contains(s, `"tab":"Cards (50)"`) {
		t.Errorf("want no auto-open and one count on the tab, got %s", s)
	}
}
