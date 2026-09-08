chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  if (request.action === "ping") {
    fetch("http://127.0.0.1:9090/upload", { method: "OPTIONS" })
      .then(() => sendResponse({ online: true }))
      .catch(() => sendResponse({ online: false }));
    return true;
  }

  if (request.action === "upload") {
    const binary = atob(request.fileBase64);
    const bytes = new Uint8Array(binary.length);
    for (let i = 0; i < binary.length; i++) {
      bytes[i] = binary.charCodeAt(i);
    }
    const blob = new Blob([bytes], { type: request.mimeType });
    const formData = new FormData();
    formData.append("file", blob, request.fileName);

    fetch("http://127.0.0.1:9090/upload", {
      method: "POST",
      body: formData
    })
    .then(res => res.json())
    .then(data => sendResponse(data))
    .catch(err => sendResponse({ error: err.message }));

    return true;
  }
});
