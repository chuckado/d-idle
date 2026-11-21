# boardorder scenarios, builder long lane, 2026-11-20 (CHEATED setups: mechanics only)

- First runs (assertions as merged in d8ef8d2e5): `board.two-accepted-chunks` Sorceress 3 GREEN, Paladin 11 RED (item_in_pack bks, a wrong assertion: the chain reads the scroll on); `board.stop-parked-chunk` RED both (the audit line is a std-log line, not a bot-log JSON line, so log_line never sees it).
- The cancel reply was HTTP 200 "stopped" in the taps log, and the stopped chunk never started (V; without the stop it starts).
- Assertions fixed in this branch (stones must not start after the stop; scroll not parked). Second round: 4 scenarios x (Paladin 11, Sorceress 3) GREEN, repeated once more GREEN (8 of 8, same verdicts: determinism pair).
