export async function copyText(value: string): Promise<void> {
  if (navigator.clipboard && window.isSecureContext) {
    try {
      await navigator.clipboard.writeText(value);
      return;
    } catch {
      // Fall through to the legacy path (permission denied, iframe policy, etc.).
    }
  }

  const field = document.createElement("textarea");
  field.value = value;
  field.setAttribute("readonly", "");
  field.style.position = "fixed";
  field.style.top = "0";
  field.style.left = "-9999px";
  field.style.opacity = "0";
  document.body.append(field);
  field.select();

  try {
    if (!document.execCommand("copy")) {
      throw new Error("Copy failed");
    }
  } finally {
    field.remove();
  }
}
