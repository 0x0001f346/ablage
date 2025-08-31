(() => {
  "use strict";

  const DEFAULT_DROPZONE_TEXT = "Drag & drop files here or click to select";
  const state = {
    config: null,
    files: {},
    ui: {},
    errorTimeout: null,
  };

  // ===== app ==============================

  async function appInit() {
    uiBuildElements();
    uiCacheElements();
    uiBindEvents();

    await configLoad();
    appUpdate();

    setInterval(appUpdate, 5 * 1000);
    setInterval(configLoad, 60 * 1000);
  }

  async function appUpdate() {
    if (state.config === null) {
      return;
    }

    uiUpdate();
    fileListFetch();
  }

  // ===== config ===========================

  async function configLoad() {
    try {
      const res = await fetch("/config/", { cache: "no-store" });
      if (!res.ok) {
        console.error("HTTP error:", res.status);
      }
      state.config = await res.json();
    } catch (err) {
      console.error("Failed to load config:", err);
      state.config = null;
    }
  }

  // ===== files ============================

  async function fileDeleteClickHandler(event, file) {
    event.preventDefault();
    if (!confirm(`Do you really want to delete "${file.Name}"?`)) return;

    try {
      const res = await fetch(
        state.config.Endpoints.FilesDelete.replace(
          ":filename",
          encodeURIComponent(file.Name)
        ),
        { method: "GET" }
      );

      if (res.ok) {
        uiShowSuccess("File deleted: " + file.Name);
      } else {
        uiShowError("Delete failed");
      }

      fileListFetch();
    } catch (err) {
      uiShowError("Delete failed");
    }
  }

  async function fileListFetch() {
    if (state.config.Modes.Sinkhole) {
      fileListClear();
      return;
    }

    try {
      const files = await fileListRequest();
      state.files = {};
      fileListClear();
      fileListRender(files);
    } catch (err) {
      console.error("fileListFetch failed:", err);
    }
  }

  async function fileListRequest() {
    const res = await fetch(state.config.Endpoints.Files, {
      cache: "no-store",
    });
    if (!res.ok) throw new Error("HTTP " + res.status);
    return res.json();
  }

  function fileListClear() {
    if (state.ui.fileList) state.ui.fileList.innerHTML = "";
  }

  function fileListRender(files) {
    if (!state.ui.fileList) return;

    files.forEach((file) => {
      state.files[file.Name] = true;

      const li = document.createElement("li");
      li.appendChild(uiCreateDownloadLink(file));

      if (!state.config.Modes.Readonly) {
        li.appendChild(uiCreateDeleteLink(file));
      }

      state.ui.fileList.appendChild(li);
    });
  }

  function fileSanitizeName(dirtyFilename) {
    if (!dirtyFilename || dirtyFilename.trim() === "") {
      return "upload.bin";
    }

    const filenameWithoutPath = dirtyFilename.split(/[\\/]/).pop();

    const lastDot = filenameWithoutPath.lastIndexOf(".");
    const extension = lastDot !== -1 ? filenameWithoutPath.slice(lastDot) : "";
    let nameOnly =
      lastDot !== -1
        ? filenameWithoutPath.slice(0, lastDot)
        : filenameWithoutPath;

    const charMap = {
      Ä: "Ae",
      ä: "ae",
      Ö: "Oe",
      ö: "oe",
      Ü: "Ue",
      ü: "ue",
      ß: "ss",
    };

    let cleanedFilename = nameOnly.replace(/./g, (char) => {
      if (charMap[char]) {
        return charMap[char];
      }
      if (char === " ") {
        return "_";
      }
      return char;
    });

    cleanedFilename = cleanedFilename.replace(/[^a-zA-Z0-9._-]+/g, "_");

    while (cleanedFilename.includes("__")) {
      cleanedFilename = cleanedFilename.replace(/__+/g, "_");
    }

    cleanedFilename = cleanedFilename.replace(/^_+|_+$/g, "");

    const MAX_LEN = 128;
    if (cleanedFilename.length > MAX_LEN) {
      cleanedFilename = cleanedFilename.slice(0, MAX_LEN);
    }

    return cleanedFilename + extension;
  }

  function fileValidateBeforeUpload(files) {
    for (const f of files) {
      const safeName = fileSanitizeName(f.name);
      if (safeName === ".upload") {
        uiShowError("Invalid filename: .upload");
        return false;
      }
      if (safeName in state.files) {
        uiShowError("File already exists: " + f.name);
        return false;
      }
    }
    return true;
  }

  // ===== ui ===============================

  function uiBindEvents() {
    state.ui.dropzone.addEventListener("click", () =>
      state.ui.fileInput.click()
    );
    state.ui.fileInput.addEventListener("change", () => {
      if (state.ui.fileInput.files.length > 0)
        uploadStart(state.ui.fileInput.files);
    });
    state.ui.dropzone.addEventListener("dragover", (e) => {
      e.preventDefault();
      state.ui.dropzone.style.borderColor = "#0fff50";
    });
    state.ui.dropzone.addEventListener("dragleave", () => {
      state.ui.dropzone.style.borderColor = "#888";
    });
    state.ui.dropzone.addEventListener("drop", (e) => {
      e.preventDefault();
      state.ui.dropzone.style.borderColor = "#888";
      if (e.dataTransfer.files.length > 0) uploadStart(e.dataTransfer.files);
    });
  }

  function uiBuildElements() {
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
    divDropzone.innerHTML = DEFAULT_DROPZONE_TEXT;
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

  function uiCacheElements() {
    state.ui.currentFileName = document.getElementById("currentFileName");
    state.ui.dropzone = document.getElementById("dropzone");
    state.ui.fileInput = document.getElementById("fileInput");
    state.ui.fileList = document.getElementById("file-list");
    state.ui.overallProgress = document.getElementById("overallProgress");
    state.ui.overallStatus = document.getElementById("overallStatus");
    state.ui.overallProgressContainer = document.getElementById(
      "overallProgressContainer"
    );
    state.ui.sinkholeModeInfo = document.getElementById("sinkholeModeInfo");
  }

  function uiCreateDeleteLink(file) {
    const link = document.createElement("a");
    link.className = "delete-link";
    link.href = "#";
    link.textContent = " [Delete]";
    link.title = "Delete file";
    link.addEventListener("click", (e) => fileDeleteClickHandler(e, file));
    return link;
  }

  function uiCreateDownloadLink(file) {
    const size = uiFormatSize(file.Size);
    const link = document.createElement("a");
    link.className = "download-link";
    link.href = state.config.Endpoints.FilesGet.replace(
      ":filename",
      encodeURIComponent(file.Name)
    );
    link.textContent = `${file.Name} (${size})`;
    return link;
  }

  function uiFormatSize(bytes) {
    const units = ["B", "KB", "MB", "GB", "TB"];
    let i = 0;
    while (bytes >= 1024 && i < units.length - 1) {
      bytes /= 1024;
      i++;
    }
    return `${bytes.toFixed(1)} ${units[i]}`;
  }

  function uiFormatSpeed(bytesPerSec) {
    if (!isFinite(bytesPerSec) || bytesPerSec <= 0) return "—";
    if (bytesPerSec < 1024) return bytesPerSec.toFixed(0) + " B/s";
    if (bytesPerSec < 1024 * 1024)
      return (bytesPerSec / 1024).toFixed(1) + " KB/s";
    return (bytesPerSec / (1024 * 1024)).toFixed(2) + " MB/s";
  }

  function uiInitProgress() {
    state.ui.overallProgressContainer.style.display = "block";
    state.ui.overallProgress.value = 0;
    state.ui.overallStatus.textContent = "";
    state.ui.currentFileName.textContent = "";
  }

  function uiShowError(msg) {
    uiShowMessage(msg, "error", 2000);
  }

  function uiShowMessage(msg, type, duration = 2000) {
    state.ui.dropzone.innerHTML = msg;
    state.ui.dropzone.classList.add(type);

    if (state.errorTimeout) clearTimeout(state.errorTimeout);

    state.errorTimeout = setTimeout(() => {
      state.ui.dropzone.innerHTML = DEFAULT_DROPZONE_TEXT;
      state.ui.dropzone.classList.remove(type);
      state.errorTimeout = null;
    }, duration);
  }

  function uiShowSuccess(msg) {
    uiShowMessage(msg, "success", 1500);
  }

  function uiUpdate() {
    if (state.config.Modes.Readonly) {
      state.ui.dropzone.style.display = "none";
    } else {
      state.ui.dropzone.style.display = "block";
    }

    if (state.config.Modes.Sinkhole) {
      state.ui.fileList.style.display = "none";
      state.ui.sinkholeModeInfo.style.display = "block";
    } else {
      state.ui.fileList.style.display = "block";
      state.ui.sinkholeModeInfo.style.display = "none";
    }
  }

  function uiUpdateProgress(totalUploaded, totalSize, startTime) {
    const percent = (totalUploaded / totalSize) * 100;
    state.ui.overallProgress.value = percent;

    const elapsed = (Date.now() - startTime) / 1000;
    const speed = totalUploaded / elapsed;
    const speedStr = uiFormatSpeed(speed);

    const remainingBytes = totalSize - totalUploaded;
    const etaSec = speed > 0 ? remainingBytes / speed : Infinity;
    const min = Math.floor(etaSec / 60);
    const sec = Math.floor(etaSec % 60);

    state.ui.overallStatus.textContent =
      `${percent.toFixed(1)}% (${(totalSize / 1024 / 1024).toFixed(
        1
      )} MB total) — ` +
      `Speed: ${speedStr}, Est. time left: ${
        isFinite(etaSec) ? `${min}m ${sec}s` : "calculating…"
      }`;
  }

  // ===== upload ===========================

  function uploadFinish(success) {
    state.ui.overallProgressContainer.style.display = "none";
    state.ui.overallProgress.value = 0;
    state.ui.overallStatus.textContent = "";
    state.ui.currentFileName.textContent = "";
    fileListFetch();
    if (success) {
      uiShowSuccess("Upload successful");
    }
  }

  function uploadStart(fileListLike) {
    const files = Array.from(fileListLike);
    if (files.length === 0) return;

    if (!fileValidateBeforeUpload(files)) return;

    uiInitProgress();

    const totalSize = files.reduce((sum, f) => sum + f.size, 0);
    let uploadedBytes = 0;
    let currentIndex = 0;
    const startTime = Date.now();
    let allSuccessful = true;

    function uploadNext() {
      if (currentIndex >= files.length) {
        uploadFinish(allSuccessful);
        return;
      }

      const file = files[currentIndex];
      state.ui.currentFileName.textContent = file.name;

      const xhr = new XMLHttpRequest();
      const form = new FormData();
      form.append("uploadfile", file);

      xhr.upload.addEventListener("progress", (e) => {
        if (e.lengthComputable) {
          uiUpdateProgress(uploadedBytes + e.loaded, totalSize, startTime);
        }
      });

      xhr.addEventListener("load", () => {
        if (xhr.status === 200) {
          uploadedBytes += file.size;
        } else if (xhr.status === 409) {
          uiShowError("File already exists: " + file.name);
          allSuccessful = false;
        } else {
          uiShowError("Upload failed: " + file.name);
          allSuccessful = false;
        }
        currentIndex++;
        uploadNext();
      });

      xhr.addEventListener("error", () => {
        uiShowError("Network or server error during upload.");
        allSuccessful = false;
        currentIndex++;
        uploadNext();
      });

      xhr.open("POST", state.config.Endpoints.Upload);
      xhr.send(form);
    }

    fileListFetch();
    uploadNext();
  }

  // ===== init ============================

  document.addEventListener("DOMContentLoaded", appInit);
})();
