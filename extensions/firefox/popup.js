document.addEventListener("DOMContentLoaded", async () => {
  const hostStatus = document.getElementById("hostStatus");
  const toggleInterception = document.getElementById("toggleInterception");
  const btnPrompt = document.getElementById("btnPrompt");
  const btnAuto = document.getElementById("btnAuto");

  // 1. Check local interception toggle
  const storage = await browser.storage.local.get("enabled");
  toggleInterception.checked = storage.enabled !== false;

  toggleInterception.addEventListener("change", async () => {
    await browser.storage.local.set({ enabled: toggleInterception.checked });
  });

  // 2. Query Downstream Desktop for connection & settings
  async function syncSettings() {
    try {
      const response = await browser.runtime.sendNativeMessage("downstream", {
        action: "get_settings",
      });

      if (response && response.status === "ok") {
        hostStatus.textContent = "Connected";
        hostStatus.className = "status-badge online";

        const action = response.settings?.BrowserDownloadAction || "prompt";
        updateActionButtons(action);
      } else {
        hostStatus.textContent = "Offline";
        hostStatus.className = "status-badge offline";
      }
    } catch (_) {
      hostStatus.textContent = "Disconnected";
      hostStatus.className = "status-badge offline";
    }
  }

  function updateActionButtons(action) {
    if (action === "auto") {
      btnAuto.classList.add("active");
      btnPrompt.classList.remove("active");
    } else {
      btnPrompt.classList.add("active");
      btnAuto.classList.remove("active");
    }
  }

  // 3. Update settings when clicked
  async function setAction(action) {
    updateActionButtons(action);
    try {
      await browser.runtime.sendNativeMessage("downstream", {
        action: "update_settings",
        key: "browser_download_action",
        value: action,
      });
    } catch (err) {
      console.warn("Failed to update setting:", err);
    }
  }

  btnPrompt.addEventListener("click", () => setAction("prompt"));
  btnAuto.addEventListener("click", () => setAction("auto"));

  await syncSettings();
});
