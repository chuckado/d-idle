import sys,collections
d=collections.defaultdict(dict)
for l in open('/tmp/d28_game.tsv'):
    f=l.split()
    if f[0]!='G':continue
    sid,st,ev,cm,nm,v=f[1],f[2],f[3],int(f[4]),int(f[5]),int(f[6])
    d[(sid,st,ev,cm)][nm]=v
for k in sorted(d):
    if k[2]!='live' and k[1]!='none': continue
    al=[n for n,v in sorted(d[k].items()) if v]
    print(k, al)
