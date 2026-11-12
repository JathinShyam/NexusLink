(function () {
  try {
    var stored = localStorage.getItem("nexuslink.theme");
    var theme =
      stored === "dark" || stored === "light"
        ? stored
        : window.matchMedia("(prefers-color-scheme: dark)").matches
          ? "dark"
          : "light";
    document.documentElement.dataset.theme = theme;
    var meta = document.querySelector('meta[name="theme-color"]');
    if (meta) {
      meta.setAttribute("content", theme === "dark" ? "#0b0c0e" : "#f2f2ed");
    }
  } catch {
    /* Preferences are optional. */
  }
})();
