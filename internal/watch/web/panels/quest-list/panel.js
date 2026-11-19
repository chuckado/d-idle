// The phone's Quests tab (design 4.1, 4.3): the board list as a full-screen tab under the strip tile, no sheet. The same list as
// the quest-board drawer (panels/quest-board/panel.js); on a desktop layout this panel is not placed.
import { mountBoard } from '../quest-board/panel.js';

export function mount(root, ctx) { return mountBoard(root, ctx, true); }
