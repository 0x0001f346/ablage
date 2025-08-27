(() => {
  "use strict";

  let AppConfig = null;
  let UI = {};

  async function appLoop() {
    if (AppConfig === null) {
      return;
    }

    updateUI();
    fetchFiles();
  }

  function updateUI() {
    if (AppConfig.Modes.Readonly) {
      UI.dropzone.style.display = "none";
    } else {
      UI.dropzone.style.display = "block";
    }

    if (AppConfig.Modes.Sinkhole) {
      UI.fileList.style.display = "none";
      UI.sinkholeModeInfo.style.display = "block";
    } else {
      UI.fileList.style.display = "block";
      UI.sinkholeModeInfo.style.display = "none";
    }
  }

  async function initApp() {
    addUIElementsToBody();
    getUIElements();
    addEventListeners();

    await loadAppConfig();
    appLoop();

    setInterval(appLoop, 5 * 1000);
    setInterval(loadAppConfig, 60 * 1000);
  }

  async function loadAppConfig() {
    try {
      const res = await fetch("/config/", { cache: "no-store" });
      if (!res.ok) {
        console.error("HTTP error:", res.status);
      }
      AppConfig = await res.json();
    } catch (err) {
      console.error("Failed to load config:", err);
      AppConfig = null;
    }
  }

  async function fetchFiles() {
    if (AppConfig.Modes.Sinkhole) {
      UI.fileList.innerHTML = "";
      return;
    }

    try {
      const res = await fetch(AppConfig.Endpoints.Files, { cache: "no-store" });
      if (!res.ok) throw new Error("HTTP " + res.status);
      const files = await res.json();

      if (!UI.fileList) return;
      UI.fileList.innerHTML = "";
      files.forEach((file) => {
        const size = humanReadableSize(file.Size);

        const li = document.createElement("li");

        const downloadLink = document.createElement("a");
        downloadLink.className = "download-link";
        downloadLink.href = AppConfig.Endpoints.FilesGet.replace(
          ":filename",
          encodeURIComponent(file.Name)
        );
        downloadLink.textContent = `${file.Name} (${size})`;

        li.appendChild(downloadLink);

        if (!AppConfig.Modes.Readonly) {
          const deleteLink = document.createElement("a");
          deleteLink.className = "delete-link";
          deleteLink.href = "#";
          deleteLink.textContent = " [Delete]";
          deleteLink.title = "Delete file";
          deleteLink.addEventListener("click", async (e) => {
            e.preventDefault();
            if (!confirm(`Do you really want to delete "${file.Name}"?`))
              return;
            try {
              const r = await fetch(
                AppConfig.Endpoints.FilesDelete.replace(
                  ":filename",
                  encodeURIComponent(file.Name)
                ),
                { method: "GET" }
              );
              if (!r.ok) throw new Error("Delete failed " + r.status);
              fetchFiles();
            } catch (err) {
              console.error(err);
            }
          });

          li.appendChild(deleteLink);
        }

        UI.fileList.appendChild(li);
      });
    } catch (err) {
      console.error("fetchFiles failed:", err);
    }
  }

  function addEventListeners() {
    UI.dropzone.addEventListener("click", () => UI.fileInput.click());
    UI.fileInput.addEventListener("change", () => {
      if (UI.fileInput.files.length > 0) uploadFiles(UI.fileInput.files);
    });
    UI.dropzone.addEventListener("dragover", (e) => {
      e.preventDefault();
      UI.dropzone.style.borderColor = "#0fff50";
    });
    UI.dropzone.addEventListener("dragleave", () => {
      UI.dropzone.style.borderColor = "#888";
    });
    UI.dropzone.addEventListener("drop", (e) => {
      e.preventDefault();
      UI.dropzone.style.borderColor = "#888";
      if (e.dataTransfer.files.length > 0) uploadFiles(e.dataTransfer.files);
    });
  }

  function addUIElementsToBody() {
    document.body.innerHTML = "";

    const aLogo = document.createElement("a");
    aLogo.href = "/";
    aLogo.className = "logo";
    const h1Logo = document.createElement("h1");
    h1Logo.textContent = "Ablage";
    aLogo.appendChild(h1Logo);
    document.body.appendChild(aLogo);

    const divDropzone = document.createElement("div");
    divDropzone.id = "dropzone";
    divDropzone.innerHTML = "Drag & drop files here or click to select";
    divDropzone.style.display = "none";
    document.body.appendChild(divDropzone);

    const fileInput = document.createElement("input");
    fileInput.type = "file";
    fileInput.id = "fileInput";
    fileInput.name = "uploadfile";
    fileInput.multiple = true;
    fileInput.style.display = "none";
    document.body.appendChild(fileInput);

    const divOverallProgressContainer = document.createElement("div");
    divOverallProgressContainer.id = "overallProgressContainer";
    divOverallProgressContainer.style.display = "none";
    const divCurrentFileName = document.createElement("div");
    divCurrentFileName.id = "currentFileName";
    const progressOverall = document.createElement("progress");
    progressOverall.id = "overallProgress";
    progressOverall.value = 0;
    progressOverall.max = 100;
    const divOverallStatus = document.createElement("div");
    divOverallStatus.id = "overallStatus";
    divOverallStatus.className = "status";
    divOverallProgressContainer.appendChild(divCurrentFileName);
    divOverallProgressContainer.appendChild(progressOverall);
    divOverallProgressContainer.appendChild(divOverallStatus);
    document.body.appendChild(divOverallProgressContainer);

    const ulFileList = document.createElement("ul");
    ulFileList.id = "file-list";
    document.body.appendChild(ulFileList);

    const divSinkholeModeInfo = document.createElement("div");
    divSinkholeModeInfo.id = "sinkholeModeInfo";
    divSinkholeModeInfo.className = "sinkholeModeInfo";
    divSinkholeModeInfo.style.display = "none";
    divSinkholeModeInfo.textContent =
      "- Sinkhole mode enabled, no files will get listed -";
    document.body.appendChild(divSinkholeModeInfo);
  }

  function getUIElements() {
    UI.currentFileName = document.getElementById("currentFileName");
    UI.dropzone = document.getElementById("dropzone");
    UI.fileInput = document.getElementById("fileInput");
    UI.fileList = document.getElementById("file-list");
    UI.overallProgress = document.getElementById("overallProgress");
    UI.overallStatus = document.getElementById("overallStatus");
    UI.overallProgressContainer = document.getElementById(
      "overallProgressContainer"
    );
    UI.sinkholeModeInfo = document.getElementById("sinkholeModeInfo");
  }

  function humanReadableSize(bytes) {
    const units = ["B", "KB", "MB", "GB", "TB"];
    let i = 0;
    while (bytes >= 1024 && i < units.length - 1) {
      bytes /= 1024;
      i++;
    }
    return `${bytes.toFixed(1)} ${units[i]}`;
  }

  function humanReadableSpeed(bytesPerSec) {
    if (!isFinite(bytesPerSec) || bytesPerSec <= 0) return "—";
    if (bytesPerSec < 1024) return bytesPerSec.toFixed(0) + " B/s";
    if (bytesPerSec < 1024 * 1024)
      return (bytesPerSec / 1024).toFixed(1) + " KB/s";
    return (bytesPerSec / (1024 * 1024)).toFixed(2) + " MB/s";
  }

  function uploadFiles(fileListLike) {
    const files = Array.from(fileListLike);
    if (files.length === 0) return;

    UI.overallProgressContainer.style.display = "block";
    UI.overallProgress.value = 0;
    UI.overallStatus.textContent = "";
    UI.currentFileName.textContent = "";

    const totalSize = files.reduce((sum, f) => sum + f.size, 0);
    let uploadedBytes = 0;
    const t0 = Date.now();
    let idx = 0;

    const uploadNext = () => {
      if (idx >= files.length) {
        UI.overallProgressContainer.style.display = "none";
        UI.overallProgress.value = 0;
        UI.overallStatus.textContent = "";
        UI.currentFileName.textContent = "";
        fetchFiles();
        return;
      }

      const file = files[idx];
      UI.currentFileName.textContent = file.name;

      const xhr = new XMLHttpRequest();
      const form = new FormData();
      form.append("uploadfile", file);

      xhr.upload.addEventListener("progress", (e) => {
        if (!e.lengthComputable) return;

        const totalUploaded = uploadedBytes + e.loaded;
        const percent = (totalUploaded / totalSize) * 100;
        UI.overallProgress.value = percent;

        const elapsed = (Date.now() - t0) / 1000;
        const speed = totalUploaded / elapsed;
        const speedStr = humanReadableSpeed(speed);

        const remainingBytes = totalSize - totalUploaded;
        const etaSec = speed > 0 ? remainingBytes / speed : Infinity;
        const min = Math.floor(etaSec / 60);
        const sec = Math.floor(etaSec % 60);

        UI.overallStatus.textContent =
          `${percent.toFixed(1)}% (${(totalSize / 1024 / 1024).toFixed(
            1
          )} MB total) — ` +
          `Speed: ${speedStr}, Est. time left: ${
            isFinite(etaSec) ? `${min}m ${sec}s` : "calculating…"
          }`;
      });

      xhr.addEventListener("load", () => {
        if (xhr.status === 200) {
          uploadedBytes += file.size;
        } else {
          console.error("Upload failed with status", xhr.status);
        }
        idx++;
        uploadNext();
      });

      xhr.addEventListener("error", () => {
        console.error("Network/server error during upload.");
        idx++;
        uploadNext();
      });

      xhr.open("POST", AppConfig.Endpoints.Upload);
      xhr.send(form);
    };

    fetchFiles();
    uploadNext();
  }

  document.addEventListener("DOMContentLoaded", initApp);
})();
