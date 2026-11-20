import emu, csv, io
raw = emu.mpq_cache().read('data\\global\\excel\\monstats.txt').decode('latin-1')
r = list(csv.reader(io.StringIO(raw), delimiter='\t'))
h = r[0]
print([ (i, c) for i, c in enumerate(h) if 'hld' in c or 'lock' in c])
for row in r[1:]:
    if row and row[0] in ('skeleton1', 'fallen1', 'zombie1', 'mephisto', 'andariel', 'unraveler1', 'diablo'):
        print(row[0], [(c, row[i]) for i, c in enumerate(h) if 'hld' in c or 'lock' in c])
print('count nonzero', sum(1 for row in r[1:] if row and len(row) > h.index('NoShldBlock') and row[h.index('NoShldBlock')] not in ('', '0')))
