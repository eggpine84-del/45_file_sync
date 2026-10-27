let appConfig = { active_preset_id: "", presets: [] };
let activePreset = null;
let syncCopiedCount = 0;
let syncCopiedBuffer = [];
let syncSummaryShown = false;

function resetSyncLogState() {
  syncCopiedCount = 0;
  syncCopiedBuffer = [];
  syncSummaryShown = false;
}

// 초기화
window.addEventListener("DOMContentLoaded", async () => {
  await reloadConfig();

  // Wails 실시간 이벤트 수신 등록
  if (window.runtime && window.runtime.EventsOn) {
    window.runtime.EventsOn("sync:progress", (data) => {
      if (data.message && data.message.trim() !== "") {
        if (data.message.startsWith("[COPIED]")) {
          syncCopiedCount++;
          if (syncCopiedCount <= 3) {
            appendLog(data.message);
          } else if (syncCopiedCount <= 5) {
            syncCopiedBuffer.push(data.message);
          } else if (!syncSummaryShown) {
            syncSummaryShown = true;
            syncCopiedBuffer = [];
            const remaining = data.total > 0 && data.total > syncCopiedCount ? data.total - 3 : syncCopiedCount - 3;
            appendLog(`[INFO] 외 ${remaining}개 파일 추가 복사 중...`, "zinc");
          }
        } else {
          appendLog(data.message);
        }
      }
      if (data.total > 0) {
        const pct = Math.round((data.current / data.total) * 100);
        updateProgressBar(pct, `${data.current} / ${data.total} 파일 동기화 중...`);
      }
    });

    window.runtime.EventsOn("sync:completed", (data) => {
      const runBtn = document.getElementById("runSyncBtn");
      runBtn.disabled = false;
      runBtn.classList.remove("opacity-50");

      if (data.success) {
        updateProgressBar(100, "동기화 완료!");
        const copied = data.result.total_copied || 0;
        const skipped = data.result.total_skipped || 0;
        const errors = data.result.total_errors || 0;

        if (syncCopiedCount <= 5 && syncCopiedBuffer.length > 0) {
          syncCopiedBuffer.forEach((msg) => appendLog(msg));
        }

        if (skipped > 0 && copied === 0 && errors === 0) {
          appendLog(`[INFO] ${skipped}개 항목 변경 없음 (모두 최신 상태)`, "zinc");
        } else if (skipped > 0) {
          appendLog(`[INFO] 나머지 ${skipped}개 항목 변경 없음`, "zinc");
        }
        appendLog(`[DONE] 동기화 성공: 복사 ${copied}건, 스킵 ${skipped}건, 오류 ${errors}건`, "emerald");
      } else {
        appendLog(`[FAIL] 동기화 실패: ${data.error}`, "red");
        updateProgressBar(0, "오류로 중단됨");
      }
      resetSyncLogState();
    });
  }
});

async function reloadConfig() {
  if (!window.go || !window.go.app || !window.go.app.App) return;
  const res = await window.go.app.App.GetConfig();
  if (res && res.success) {
    appConfig = res.data;
    renderPresetList();
    selectActivePreset(appConfig.active_preset_id);
  }
}


async function switchPreset(id) {
  if (window.go && window.go.app && window.go.app.App) {
    await window.go.app.App.SetActivePreset(id);
  }
  selectActivePreset(id);
  renderPresetList();
}


async function createNewPreset() {
  const defaultName = "신규 프리셋 " + (appConfig.presets.length + 1);
  const inputName = prompt("새 프리셋 이름을 입력하세요:", defaultName);
  if (inputName === null) return; // 취소 클릭 시 생성 중단
  const finalName = inputName.trim() || defaultName;

  const newPreset = {
    id: "preset-" + Date.now(),
    name: finalName,
    sources: [],
    targets: [],
    auto_watch: false
  };
  if (window.go && window.go.app && window.go.app.App) {
    const res = await window.go.app.App.SavePreset(newPreset);
    if (res && res.success) {
      appConfig = res.data;
      selectActivePreset(newPreset.id);
      renderPresetList();
      appendLog(`[PRESET] 새 프리셋 생성: ${newPreset.name}`);
    }
  }
}

function handlePresetNameInput(newName) {
  if (!activePreset) return;
  activePreset.name = newName;
  renderPresetList();
}

async function updatePresetName(newName) {
  if (!activePreset) return;
  const trimmed = newName.trim();
  if (!trimmed) {
    document.getElementById("presetNameInput").value = activePreset.name;
    return;
  }
  activePreset.name = trimmed;
  await saveActivePreset();
  renderPresetList();
}

async function addSourceFiles() {
  if (!activePreset) return;
  if (window.go && window.go.app && window.go.app.App) {
    const res = await window.go.app.App.SelectFiles();
    if (res && res.success && res.data && res.data.length > 0) {
      activePreset.sources = activePreset.sources || [];
      res.data.forEach(f => {
        if (!activePreset.sources.includes(f)) activePreset.sources.push(f);
      });
      await saveActivePreset();
      renderPresetDetails();
    }
  }
}

async function addSourceDirectory() {
  if (!activePreset) return;
  if (window.go && window.go.app && window.go.app.App) {
    const res = await window.go.app.App.SelectDirectory();
    if (res && res.success && res.data) {
      activePreset.sources = activePreset.sources || [];
      if (!activePreset.sources.includes(res.data)) activePreset.sources.push(res.data);
      await saveActivePreset();
      renderPresetDetails();
    }
  }
}

async function addTargetDirectory() {
  if (!activePreset) return;
  if (window.go && window.go.app && window.go.app.App) {
    const res = await window.go.app.App.SelectDirectory();
    if (res && res.success && res.data) {
      activePreset.targets = activePreset.targets || [];
      if (!activePreset.targets.includes(res.data)) activePreset.targets.push(res.data);
      await saveActivePreset();
      renderPresetDetails();
    }
  }
}

async function removeSource(index) {
  if (!activePreset || !activePreset.sources) return;
  activePreset.sources.splice(index, 1);
  await saveActivePreset();
  renderPresetDetails();
}

async function removeTarget(index) {
  if (!activePreset || !activePreset.targets) return;
  activePreset.targets.splice(index, 1);
  await saveActivePreset();
  renderPresetDetails();
}

async function deleteCurrentPreset() {
  if (!activePreset) return;
  if (!confirm(`정말 '${activePreset.name}' 프리셋을 삭제하시겠습니까?`)) return;

  if (window.go && window.go.app && window.go.app.App) {
    const res = await window.go.app.App.DeletePreset(activePreset.id);
    if (res && res.success) {
      appConfig = res.data;
      selectActivePreset(appConfig.active_preset_id);
      renderPresetList();
      appendLog(`[PRESET] 프리셋 삭제 완료`);
    }
  }
}

async function saveActivePreset() {
  if (!activePreset) return;
  if (window.go && window.go.app && window.go.app.App) {
    const res = await window.go.app.App.SavePreset(activePreset);
    if (res && res.success) {
      appConfig = res.data;
    }
  }
}

async function runSync() {
  if (!activePreset) {
    alert("선택된 프리셋이 없습니다.");
    return;
  }
  if (!activePreset.sources || activePreset.sources.length === 0) {
    alert("동기화할 소스(파일/폴더)를 최소 1개 이상 추가하세요.");
    return;
  }
  if (!activePreset.targets || activePreset.targets.length === 0) {
    alert("동기화 대상 폴더를 최소 1개 이상 추가하세요.");
    return;
  }

  const runBtn = document.getElementById("runSyncBtn");
  runBtn.disabled = true;
  runBtn.classList.add("opacity-50");
  resetSyncLogState();
  updateProgressBar(5, "동기화 시작 중...");
  appendLog(`[START] '${activePreset.name}' 전체 동기화 시작...`);

  if (window.go && window.go.app && window.go.app.App) {
    const res = await window.go.app.App.RunSync(activePreset.id);
    if (res && !res.success) {
      runBtn.disabled = false;
      runBtn.classList.remove("opacity-50");
      appendLog(`[ERROR] 시작 실패: ${res.error}`, "red");
      updateProgressBar(0, "실패");
    }
  }
}

async function toggleWatchDaemon(enabled) {
  if (!activePreset) return;
  activePreset.auto_watch = enabled;
  await saveActivePreset();

  if (window.go && window.go.app && window.go.app.App) {
    const res = await window.go.app.App.ToggleWatch(enabled);
    if (res && res.success) {
      appendLog(enabled ? "[WATCH] 실시간 감시 데몬 가동" : "[WATCH] 실시간 감시 중지");
    } else if (res && !res.success) {
      appendLog(`[WATCH ERROR] ${res.error}`, "red");
      document.getElementById("watchToggle").checked = !enabled;
    }
  }
}

function hideToTray() {
  if (window.go && window.go.app && window.go.app.App) {
    window.go.app.App.HideWindow();
  }
}

