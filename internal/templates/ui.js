// [ui.js] 순수 UI 렌더링, DOM 조작 및 콘솔 로그 유틸리티 모듈

function renderPresetList() {
  const listEl = document.getElementById("presetList");
  listEl.innerHTML = "";

  if (!appConfig.presets || appConfig.presets.length === 0) {
    listEl.innerHTML = `<span class="text-xs text-zinc-500 block p-2">등록된 프리셋이 없습니다.</span>`;
    return;
  }

  appConfig.presets.forEach((p) => {
    const isActive = p.id === appConfig.active_preset_id;
    const item = document.createElement("button");
    item.className = `w-full text-left px-3 py-2 rounded-lg text-xs font-medium transition flex items-center justify-between ${
      isActive ? "bg-violet-600/20 text-violet-300 border border-violet-500/30" : "text-zinc-400 hover:bg-zinc-800/60 hover:text-zinc-200"
    }`;
    item.onclick = () => switchPreset(p.id);
    item.innerHTML = `
      <span class="truncate">${escapeHtml(p.name || "무제 프리셋")}</span>
      <span class="text-[10px] text-zinc-500">${(p.sources ? p.sources.length : 0)} -> ${(p.targets ? p.targets.length : 0)}</span>
    `;
    listEl.appendChild(item);
  });
}

function selectActivePreset(id) {
  if (!appConfig.presets || appConfig.presets.length === 0) {
    activePreset = null;
    renderPresetDetails();
    return;
  }
  activePreset = appConfig.presets.find(p => p.id === id) || appConfig.presets[0];
  appConfig.active_preset_id = activePreset.id;
  renderPresetDetails();
}

function renderPresetDetails() {
  const nameInput = document.getElementById("presetNameInput");
  const watchToggle = document.getElementById("watchToggle");
  const sourcesContainer = document.getElementById("sourcesContainer");
  const targetsContainer = document.getElementById("targetsContainer");
  const srcBadge = document.getElementById("sourceCountBadge");
  const tgtBadge = document.getElementById("targetCountBadge");

  if (!activePreset) {
    nameInput.value = "";
    nameInput.disabled = true;
    sourcesContainer.innerHTML = `<span class="text-xs text-zinc-500 m-auto select-none">프리셋을 생성하세요.</span>`;
    targetsContainer.innerHTML = `<span class="text-xs text-zinc-500 m-auto block text-center py-10 select-none">프리셋을 생성하세요.</span>`;
    srcBadge.textContent = "0";
    tgtBadge.textContent = "0";
    return;
  }

  nameInput.disabled = false;
  nameInput.value = activePreset.name;
  watchToggle.checked = !!activePreset.auto_watch;

  // 소스 세로 행 리스트 렌더링
  const sources = activePreset.sources || [];
  srcBadge.textContent = sources.length;
  if (sources.length === 0) {
    sourcesContainer.innerHTML = `<span class="text-xs text-zinc-500 m-auto block text-center py-10 select-none">지정된 소스가 없습니다.</span>`;
  } else {
    sourcesContainer.innerHTML = "";
    sources.forEach((src, idx) => {
      const row = document.createElement("div");
      row.className = "flex items-center justify-between p-2 rounded bg-zinc-900 border border-zinc-800 text-xs hover:border-zinc-700 transition";

      const isFile = src.includes(".") && !src.endsWith("/") && !src.endsWith("\\");
      const badgeText = isFile ? "파일" : "폴더";
      const badgeClass = isFile ? "bg-amber-950/60 text-amber-400 border-amber-800/60" : "bg-sky-950/60 text-sky-400 border-sky-800/60";

      row.innerHTML = `
        <div class="flex items-center gap-2 min-w-0 flex-1 mr-2">
          <span class="text-[10px] px-1.5 py-0.5 rounded border font-medium shrink-0 ${badgeClass}">${badgeText}</span>
          <span class="break-all font-mono text-zinc-300 select-all" title="${escapeHtml(src)}">${escapeHtml(src)}</span>
        </div>
        <button onclick="removeSource(${idx})" title="삭제" class="text-zinc-500 hover:text-red-400 font-bold px-1.5 shrink-0">×</button>
      `;
      sourcesContainer.appendChild(row);
    });
  }

  // 대상 폴더 렌더링
  const targets = activePreset.targets || [];
  tgtBadge.textContent = targets.length;
  if (targets.length === 0) {
    targetsContainer.innerHTML = `<span class="text-xs text-zinc-500 m-auto block text-center py-10 select-none">지정된 대상 폴더가 없습니다.</span>`;
  } else {
    targetsContainer.innerHTML = "";
    targets.forEach((tgt, idx) => {
      const row = document.createElement("div");
      row.className = "flex items-center justify-between p-2 rounded bg-zinc-900 border border-zinc-800 text-xs hover:border-zinc-700 transition";
      row.innerHTML = `
        <div class="flex items-center gap-2 min-w-0 flex-1 mr-2">
          <span class="text-[10px] px-1.5 py-0.5 rounded border font-medium shrink-0 bg-violet-950/60 text-violet-400 border-violet-800/60">대상</span>
          <span class="break-all font-mono text-zinc-300 select-all" title="${escapeHtml(tgt)}">${escapeHtml(tgt)}</span>
        </div>
        <button onclick="removeTarget(${idx})" title="삭제" class="text-zinc-500 hover:text-red-400 font-bold px-1.5 shrink-0">×</button>
      `;
      targetsContainer.appendChild(row);
    });
  }
}

function appendLog(msg, color = "zinc") {
  const consoleEl = document.getElementById("logConsole");
  const timeStr = new Date().toTimeString().split(" ")[0];
  const line = document.createElement("div");

  let colorClass = "text-zinc-300";
  if (color === "emerald" || msg.includes("[COPIED]") || msg.includes("[DONE]")) colorClass = "text-emerald-400";
  if (color === "red" || msg.includes("[ERROR]") || msg.includes("[FAIL]")) colorClass = "text-red-400 font-semibold";
  if (msg.includes("[SKIP]")) colorClass = "text-zinc-500";
  if (msg.includes("[WATCH]")) colorClass = "text-violet-400";

  line.className = colorClass;
  line.textContent = `[${timeStr}] ${msg}`;
  consoleEl.appendChild(line);
  consoleEl.scrollTop = consoleEl.scrollHeight;
}

function clearLogs() {
  document.getElementById("logConsole").innerHTML = "";
}

function updateProgressBar(pct, summaryText) {
  document.getElementById("syncProgressBar").style.width = pct + "%";
  if (summaryText) {
    document.getElementById("syncStatusSummary").textContent = summaryText;
  }
}

function escapeHtml(str) {
  if (!str) return "";
  return String(str)
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#039;");
}
