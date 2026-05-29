function togglePasswordVisibility(button) {
  const field = button.closest(".password-field");
  const input = field ? field.querySelector("input") : null;
  if (!input) return;

  const isVisible = input.type === "text";
  input.type = isVisible ? "password" : "text";
  button.setAttribute("aria-label", isVisible ? "Show password" : "Hide password");
  button.setAttribute("aria-pressed", String(!isVisible));
}

export function initializePasswordToggles(root = document) {
  root.addEventListener(
    "click",
    (event) => {
      const target = event.target instanceof Element ? event.target : null;
      const button = target?.closest("[data-password-toggle]");
      if (!button) return;

      event.preventDefault();
      togglePasswordVisibility(button);
    },
    true
  );
}
