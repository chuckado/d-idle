# Smith stand click, why the scenario was RED (toolstrade38, 2026-11-23)

Author: toolstrade38 (internal agent)

- V lab run 20261123-194827 (devp25, Paladin 11): the board reached "Pick up Horadric Malus from the Malus stand", the step ended
  "ran but is still needed", no Malus. The loot sweep inside `taskActs.PickUp` took three other items (gold, a scroll, a buckler)
  and the step trusted the pick-up count, so the stand was never clicked. The adapter lookup and the spawned object were fine.
- Fix: `PickUp` of a named quest item is success only when the item is in the pack afterwards (internal/bot/task.go,
  TestPickUpOfAQuestItemChecksThePack, must fail without the check). Otherwise it is KindNotFound and the step clicks the stand.
- V lab run 20261123-195923 with the fix: stand clicked, `loot: clicked item=Horadric Malus`, `Found Horadric Malus`, task step
  `malus` ran, then Charsi talk done: the Malus is handed in inside the budget, so the pack assert (`hdm`) could not hold. The
  scenario now asserts the click line. Smith kill: the boss script was still cancelled by safety interrupts in that run
  (unrelated; the Smith is dead by the time the stand step runs in the scenario).
