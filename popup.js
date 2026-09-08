document.addEventListener("DOMContentLoaded", () => {
  const dropZone = document.getElementById("drop-zone");
  const fileInput = document.getElementById("file-input");
  const resultEl = document.getElementById("result");
  const statusEl = document.getElementById("status");

  chrome.runtime.sendMessage({ action: "ping" }, (res) => {
    if (res && res.online) {
      statusEl.textContent = "就绪，拖入文件即可上传";
    } else {
      statusEl.innerHTML = '服务未启动，请先双击 <b>install.bat</b>（仅需一次）';
      dropZone.style.opacity = "0.4";
      dropZone.style.pointerEvents = "none";
    }
  });

  dropZone.addEventListener("click", () => fileInput.click());
  dropZone.addEventListener("dragover", (e) => { e.preventDefault(); dropZone.classList.add("hover"); });
  dropZone.addEventListener("dragleave", () => dropZone.classList.remove("hover"));
  dropZone.addEventListener("drop", (e) => {
    e.preventDefault();
    dropZone.classList.remove("hover");
    if (e.dataTransfer.files.length > 0) handleFileUpload(e.dataTransfer.files[0]);
  });
  fileInput.addEventListener("change", (e) => {
    if (e.target.files.length > 0) handleFileUpload(e.target.files[0]);
  });

  async function handleFileUpload(file) {
    statusEl.textContent = "上传中...";
    resultEl.style.display = "none";
    const base64 = await new Promise((resolve) => {
      const reader = new FileReader();
      reader.onload = () => resolve(reader.result.split(",")[1]);
      reader.readAsDataURL(file);
    });
    chrome.runtime.sendMessage(
      { action: "upload", fileName: file.name, fileBase64: base64, mimeType: file.type },
      (res) => {
        if (chrome.runtime.lastError) { statusEl.textContent = "通信失败"; return; }
        if (res && res.error) { statusEl.textContent = res.error; return; }
        if (res && res.code === 0 && res.url) {
          navigator.clipboard.writeText(res.url);
          resultEl.textContent = res.url;
          resultEl.style.display = "block";
          statusEl.textContent = "链接已复制到剪贴板";
        } else {
          statusEl.textContent = "上传失败";
        }
      }
    );
  }
});
