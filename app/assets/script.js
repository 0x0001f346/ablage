(() => {
  "use strict";

  let AppConfig = null;
  let ErrorTimeout = null;
  let Files = {};
  const UI = {};

  async function appLoop() {
    if (AppConfig === null) {
      return;
    }

    updateUI();
    fetchFiles();
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
      clearFileList();
      return;
    }

    try {
      const files = await fetchFileList();
      Files = {};
      clearFileList();
      renderFileList(files);
    } catch (err) {
      console.error("fetchFiles failed:", err);
    }
  }

  async function fetchFileList() {
    const res = await fetch(AppConfig.Endpoints.Files, { cache: "no-store" });
    if (!res.ok) throw new Error("HTTP " + res.status);
    return res.json();
  }

  async function handleDeleteClick(event, file) {
    event.preventDefault();
    if (!confirm(`Do you really want to delete "${file.Name}"?`)) return;

    try {
      const res = await fetch(
        AppConfig.Endpoints.FilesDelete.replace(
          ":filename",
          encodeURIComponent(file.Name)
        ),
        { method: "GET" }
      );

      if (res.ok) {
        showSuccess("File deleted");
      } else {
        showError("Delete failed");
      }

      fetchFiles();
    } catch (err) {
      showError("Delete failed");
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
    divDropzone.className = "dropzone";
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

  function clearFileList() {
    if (UI.fileList) UI.fileList.innerHTML = "";
  }

  function createDownloadLink(file) {
    const size = humanReadableSize(file.Size);
    const link = document.createElement("a");
    link.className = "download-link";
    link.href = AppConfig.Endpoints.FilesGet.replace(
      ":filename",
      encodeURIComponent(file.Name)
    );
    link.textContent = `${file.Name} (${size})`;
    return link;
  }

  function createDeleteLink(file) {
    const link = document.createElement("a");
    link.className = "delete-link";
    link.href = "#";
    link.textContent = " [Delete]";
    link.title = "Delete file";
    link.addEventListener("click", (e) => handleDeleteClick(e, file));
    return link;
  }

  function finishUpload(success) {
    UI.overallProgressContainer.style.display = "none";
    UI.overallProgress.value = 0;
    UI.overallStatus.textContent = "";
    UI.currentFileName.textContent = "";
    fetchFiles();
    if (success) {
      showSuccess("Upload successful");
    }
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

  function initUIProgress() {
    UI.overallProgressContainer.style.display = "block";
    UI.overallProgress.value = 0;
    UI.overallStatus.textContent = "";
    UI.currentFileName.textContent = "";
  }

  function renderFileList(files) {
    if (!UI.fileList) return;

    files.forEach((file) => {
      Files[file.Name] = true;

      const li = document.createElement("li");
      li.appendChild(createDownloadLink(file));

      if (!AppConfig.Modes.Readonly) {
        li.appendChild(createDeleteLink(file));
      }

      UI.fileList.appendChild(li);
    });
  }

  function sanitizeFilename(dirtyFilename) {
    if (!dirtyFilename || dirtyFilename.trim() === "") {
      return "upload.bin";
    }

    const filenameWithoutPath = dirtyFilename.split(/[\\/]/).pop();

    const lastDot = filenameWithoutPath.lastIndexOf(".");
    const extension = lastDot !== -1 ? filenameWithoutPath.slice(lastDot) : "";
    let filenameWithoutPathAndExtension =
      lastDot !== -1
        ? filenameWithoutPath.slice(0, lastDot)
        : filenameWithoutPath;

    let cleanedFilename = filenameWithoutPathAndExtension
      .replace(/ /g, "_")
      .replace(/Ä/g, "Ae")
      .replace(/ä/g, "ae")
      .replace(/Ö/g, "Oe")
      .replace(/ö/g, "oe")
      .replace(/Ü/g, "Ue")
      .replace(/ü/g, "ue")
      .replace(/ß/g, "ss");

    cleanedFilename = cleanedFilename.replace(/[^a-zA-Z0-9._-]+/g, "_");

    while (cleanedFilename.includes("__")) {
      cleanedFilename = cleanedFilename.replace(/__+/g, "_");
    }

    cleanedFilename = cleanedFilename.replace(/^_+|_+$/g, "");

    const maxLenFilename = 128;
    if (cleanedFilename.length > maxLenFilename) {
      cleanedFilename = cleanedFilename.slice(0, maxLenFilename);
    }

    return cleanedFilename + extension;
  }

  function showError(msg) {
    const original = "Drag & drop files here or click to select";
    UI.dropzone.innerHTML = msg;
    UI.dropzone.classList.add("error");

    if (ErrorTimeout) clearTimeout(ErrorTimeout);

    ErrorTimeout = setTimeout(() => {
      UI.dropzone.innerHTML = original;
      UI.dropzone.classList.remove("error");
      ErrorTimeout = null;
    }, 2000);
  }

  function showSuccess(msg) {
    const original = "Drag & drop files here or click to select";
    UI.dropzone.innerHTML = msg;
    UI.dropzone.classList.add("success");

    if (ErrorTimeout) clearTimeout(ErrorTimeout);

    ErrorTimeout = setTimeout(() => {
      UI.dropzone.innerHTML = original;
      UI.dropzone.classList.remove("success");
      ErrorTimeout = null;
    }, 1500);
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

  function uploadFiles(fileListLike) {
    const files = Array.from(fileListLike);
    if (files.length === 0) return;

    if (!validateFiles(files)) return;

    initUIProgress();

    const totalSize = files.reduce((sum, f) => sum + f.size, 0);
    let uploadedBytes = 0;
    let currentIndex = 0;
    const startTime = Date.now();
    let allSuccessful = true;

    function uploadNext() {
      if (currentIndex >= files.length) {
        finishUpload(allSuccessful);
        return;
      }

      const file = files[currentIndex];
      UI.currentFileName.textContent = file.name;

      const xhr = new XMLHttpRequest();
      const form = new FormData();
      form.append("uploadfile", file);

      xhr.upload.addEventListener("progress", (e) => {
        if (e.lengthComputable) {
          updateProgressUI(uploadedBytes + e.loaded, totalSize, startTime);
        }
      });

      xhr.addEventListener("load", () => {
        if (xhr.status === 200) {
          uploadedBytes += file.size;
        } else if (xhr.status === 409) {
          showError("File already exists: " + file.name);
          allSuccessful = false;
        } else {
          showError("Upload failed: " + file.name);
          allSuccessful = false;
        }
        currentIndex++;
        uploadNext();
      });

      xhr.addEventListener("error", () => {
        showError("Network or server error during upload.");
        allSuccessful = false;
        currentIndex++;
        uploadNext();
      });

      xhr.open("POST", AppConfig.Endpoints.Upload);
      xhr.send(form);
    }

    fetchFiles();
    uploadNext();
  }

  function updateProgressUI(totalUploaded, totalSize, startTime) {
    const percent = (totalUploaded / totalSize) * 100;
    UI.overallProgress.value = percent;

    const elapsed = (Date.now() - startTime) / 1000;
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
  }

  function validateFiles(files) {
    for (const f of files) {
      const safeName = sanitizeFilename(f.name);
      if (safeName === ".upload") {
        showError("Invalid filename: .upload");
        return false;
      }
      if (safeName in Files) {
        showError("File already exists: " + f.name);
        return false;
      }
    }
    return true;
  }

  document.addEventListener("DOMContentLoaded", initApp);
})();
