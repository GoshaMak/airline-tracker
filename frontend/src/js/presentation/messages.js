export function createMessageController(els) {
  let canClose = false;

  function hide() {
    els.messageOverlay.hidden = true;
    els.popupMessage.textContent = "";
    canClose = false;
  }

  function show(text, type = "") {
    if (!text) {
      hide();
      return;
    }

    els.popupMessage.textContent = text;
    els.messageDialog.className = `message-dialog ${type}`.trim();
    els.messageOverlay.hidden = false;
    canClose = false;
    window.setTimeout(() => {
      canClose = true;
    }, 0);
  }

  function closeIfReady() {
    if (!els.messageOverlay.hidden && canClose) hide();
  }

  return { closeIfReady, hide, show };
}
