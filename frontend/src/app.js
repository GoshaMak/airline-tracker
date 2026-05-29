import { createState } from "./js/application/state.js";
import { createApiClient } from "./js/infrastructure/apiClient.js";
import { clearSession, loadSession, saveSession } from "./js/infrastructure/sessionStore.js";
import { getElements } from "./js/presentation/dom.js";
import { bindEvents, refreshDataAndRender } from "./js/presentation/events.js";
import { createMessageController } from "./js/presentation/messages.js";
import { initializePasswordToggles } from "./js/presentation/passwordToggle.js";
import { renderAuthState, renderFlightViewTabs } from "./js/presentation/render.js";

const state = createState(loadSession());
const els = getElements();
const api = createApiClient(() => state.token);
const messages = createMessageController(els);
const sessionStore = { clearSession, saveSession };

initializePasswordToggles();
bindEvents({ api, els, messages, sessionStore, state });

renderAuthState(els, state);
renderFlightViewTabs(els, state);
refreshDataAndRender({ api, els, messages, sessionStore, state });
