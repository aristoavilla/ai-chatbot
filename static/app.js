/**
 * ChromaFuse AI Studio — Client Application
 * Dual-Image Creative Style Fusion Frontend Engine
 */

(function () {
  'use strict';

  // --------------------------------------------------------------------------
  // Application State
  // --------------------------------------------------------------------------
  const state = {
    imageA: null, // { dataUrl, mime, name, size }
    imageB: null, // { dataUrl, mime, name, size }
    preset: 'auto',
    intensity: 65,
    prompt: '',
    credits: 5,
    isGenerating: false,
    lastResult: null, // { imageResult, analysis, preset, durationMs }
    viewMode: 'split', // 'split' | 'side'
    splitPct: 50,
    chatHistory: [],
  };

  // --------------------------------------------------------------------------
  // DOM Elements
  // --------------------------------------------------------------------------
  // Dropzone A
  const dropAreaA = document.getElementById('drop-area-a');
  const fileInputA = document.getElementById('file-input-a');
  const dropEmptyA = document.getElementById('drop-empty-a');
  const dropPreviewA = document.getElementById('drop-preview-a');
  const previewImgA = document.getElementById('preview-img-a');
  const metaA = document.getElementById('meta-a');
  const btnRemoveA = document.getElementById('btn-remove-a');

  // Dropzone B
  const dropAreaB = document.getElementById('drop-area-b');
  const fileInputB = document.getElementById('file-input-b');
  const dropEmptyB = document.getElementById('drop-empty-b');
  const dropPreviewB = document.getElementById('drop-preview-b');
  const previewImgB = document.getElementById('preview-img-b');
  const metaB = document.getElementById('meta-b');
  const btnRemoveB = document.getElementById('btn-remove-b');

  // Controls & Presets
  const presetsCarousel = document.getElementById('presets-carousel');
  const intensitySlider = document.getElementById('intensity-slider');
  const intensityVal = document.getElementById('intensity-val');
  const customPromptInput = document.getElementById('custom-prompt');
  const btnGenerate = document.getElementById('btn-generate');
  const btnPromptGenerate = document.getElementById('btn-prompt-generate');
  const btnAiGenSubject = document.getElementById('btn-ai-gen-subject');
  const demoCatHat = document.getElementById('demo-cat-hat');
  const btnResetAll = document.getElementById('btn-reset-all');
  const creditCount = document.getElementById('credit-count');

  // Canvas Viewport States
  const canvasEmpty = document.getElementById('canvas-empty');
  const canvasLoading = document.getElementById('canvas-loading');
  const canvasResult = document.getElementById('canvas-result');
  const loadingStageTitle = document.getElementById('loading-stage-title');
  const loadingStageDesc = document.getElementById('loading-stage-desc');
  const elapsedCounter = document.getElementById('elapsed-counter');
  const progStep1 = document.getElementById('prog-step-1');
  const progStep2 = document.getElementById('prog-step-2');
  const progStep3 = document.getElementById('prog-step-3');

  // Split & Results Elements
  const splitContainer = document.getElementById('split-container');
  const splitBaseImg = document.getElementById('split-base-img');
  const splitFusedWrap = document.getElementById('split-fused-wrap');
  const splitFusedImg = document.getElementById('split-fused-img');
  const splitHandle = document.getElementById('split-handle');

  const sideContainer = document.getElementById('side-container');
  const sideImgSubject = document.getElementById('side-img-subject');
  const sideImgStyle = document.getElementById('side-img-style');
  const sideImgFused = document.getElementById('side-img-fused');

  const viewModeToggle = document.getElementById('view-mode-toggle');
  const modeSplit = document.getElementById('mode-split');
  const modeSide = document.getElementById('mode-side');
  const btnDownload = document.getElementById('btn-download');
  const btnCopyImg = document.getElementById('btn-copy-img');
  const critiqueCard = document.getElementById('critique-card');
  const critiqueBody = document.getElementById('critique-body');
  const critiqueMeta = document.getElementById('critique-meta');
  const errorBanner = document.getElementById('error-banner');

  // AI Assistant Drawer
  const btnToggleAssistant = document.getElementById('btn-toggle-assistant');
  const assistantDrawer = document.getElementById('assistant-drawer');
  const btnCloseDrawer = document.getElementById('btn-close-drawer');
  const chatMessages = document.getElementById('chat-messages');
  const chatForm = document.getElementById('chat-form');
  const userInput = document.getElementById('user-input');

  // Pricing Modal
  const btnOpenPricing = document.getElementById('btn-open-pricing');
  const pricingModal = document.getElementById('pricing-modal');
  const btnCloseModal = document.getElementById('btn-close-modal');
  const btnSimCheckout = document.getElementById('btn-sim-checkout');
  const btnSimStudio = document.getElementById('btn-sim-studio');

  // Demo Buttons
  const demoCyberpunk = document.getElementById('demo-cyberpunk');
  const demoRenaissance = document.getElementById('demo-renaissance');
  const demoGhibli = document.getElementById('demo-ghibli');

  // Toast Container
  const toastContainer = document.getElementById('toast-container');

  // --------------------------------------------------------------------------
  // Notification Toast Helper
  // --------------------------------------------------------------------------
  function showToast(message, type = 'info') {
    const toast = document.createElement('div');
    toast.className = `toast toast-${type}`;
    const icon = type === 'success' ? '✓' : type === 'error' ? '⚠' : 'ℹ';
    toast.innerHTML = `<span>${icon}</span> <span>${message}</span>`;
    toastContainer.appendChild(toast);

    setTimeout(() => {
      toast.style.opacity = '0';
      toast.style.transition = 'opacity 0.3s ease';
      setTimeout(() => toast.remove(), 300);
    }, 4000);
  }

  // --------------------------------------------------------------------------
  // Image Upload & Dropzone Logic
  // --------------------------------------------------------------------------
  function setupDropzone(dropArea, fileInput, dropEmpty, dropPreview, previewImg, metaEl, isSlotA) {
    // Click to select file
    dropArea.addEventListener('click', (e) => {
      if (e.target.closest('.btn-remove-img')) return;
      fileInput.click();
    });

    fileInput.addEventListener('change', (e) => {
      const file = e.target.files && e.target.files[0];
      if (file) handleFile(file, isSlotA);
    });

    // Drag and drop events
    ['dragenter', 'dragover'].forEach((eventName) => {
      dropArea.addEventListener(eventName, (e) => {
        e.preventDefault();
        e.stopPropagation();
        dropArea.classList.add('dragover');
      });
    });

    ['dragleave', 'drop'].forEach((eventName) => {
      dropArea.addEventListener(eventName, (e) => {
        e.preventDefault();
        e.stopPropagation();
        dropArea.classList.remove('dragover');
      });
    });

    dropArea.addEventListener('drop', (e) => {
      const dt = e.dataTransfer;
      const file = dt.files && dt.files[0];
      if (file && file.type.startsWith('image/')) {
        handleFile(file, isSlotA);
      } else {
        showToast('Please drop an image file (PNG, JPG, WebP)', 'error');
      }
    });

    // Paste listener (Ctrl+V) when hovering or focused
    window.addEventListener('paste', (e) => {
      const activeEl = document.activeElement;
      if (activeEl === userInput || activeEl === customPromptInput) return;

      const items = (e.clipboardData || e.originalEvent.clipboardData).items;
      for (const item of items) {
        if (item.type.indexOf('image') !== -1) {
          const blob = item.getAsFile();
          handleFile(blob, isSlotA);
          showToast(`Image pasted into ${isSlotA ? 'Subject' : 'Style'} slot`, 'success');
          break;
        }
      }
    });
  }

  function handleFile(file, isSlotA) {
    if (file.size > 15 * 1024 * 1024) {
      showToast('Image size exceeds 15MB limit', 'error');
      return;
    }

    const reader = new FileReader();
    reader.onload = (e) => {
      const dataUrl = e.target.result;
      const mime = file.type || 'image/jpeg';
      const sizeKb = Math.round(file.size / 1024);

      if (isSlotA) {
        state.imageA = { dataUrl, mime, name: file.name || 'subject.jpg', size: `${sizeKb} KB` };
        renderPreview(dropEmptyA, dropPreviewA, previewImgA, metaA, state.imageA);
      } else {
        state.imageB = { dataUrl, mime, name: file.name || 'style.jpg', size: `${sizeKb} KB` };
        renderPreview(dropEmptyB, dropPreviewB, previewImgB, metaB, state.imageB);
      }
    };
    reader.readAsDataURL(file);
  }

  function renderPreview(dropEmpty, dropPreview, previewImg, metaEl, imgObj) {
    previewImg.src = imgObj.dataUrl;
    metaEl.textContent = `${imgObj.name} (${imgObj.size})`;
    dropEmpty.classList.add('hidden');
    dropPreview.classList.remove('hidden');
  }

  function removeImage(isSlotA) {
    if (isSlotA) {
      state.imageA = null;
      fileInputA.value = '';
      dropPreviewA.classList.add('hidden');
      dropEmptyA.classList.remove('hidden');
    } else {
      state.imageB = null;
      fileInputB.value = '';
      dropPreviewB.classList.add('hidden');
      dropEmptyB.classList.remove('hidden');
    }
  }

  btnRemoveA.addEventListener('click', (e) => {
    e.stopPropagation();
    removeImage(true);
  });

  btnRemoveB.addEventListener('click', (e) => {
    e.stopPropagation();
    removeImage(false);
  });

  setupDropzone(dropAreaA, fileInputA, dropEmptyA, dropPreviewA, previewImgA, metaA, true);
  setupDropzone(dropAreaB, fileInputB, dropEmptyB, dropPreviewB, previewImgB, metaB, false);

  // --------------------------------------------------------------------------
  // Canvas Synthetic Demo Pair Generators (Instant 1-Click Samples)
  // --------------------------------------------------------------------------
  function createCyberpunkSubjectCanvas() {
    const c = document.createElement('canvas');
    c.width = 600;
    c.height = 600;
    const ctx = c.getContext('2d');

    // Dark sleek gradient background
    const bg = ctx.createLinearGradient(0, 0, 600, 600);
    bg.addColorStop(0, '#0f172a');
    bg.addColorStop(1, '#020617');
    ctx.fillStyle = bg;
    ctx.fillRect(0, 0, 600, 600);

    // Subject silhouette (Cyber Samurai / Protagonist)
    ctx.fillStyle = '#e2e8f0';
    // Head
    ctx.beginPath();
    ctx.arc(300, 220, 75, 0, Math.PI * 2);
    ctx.fill();

    // Body / Shoulders
    ctx.beginPath();
    ctx.ellipse(300, 420, 140, 120, 0, 0, Math.PI * 2);
    ctx.fill();

    // Subtle edge highlight
    ctx.strokeStyle = '#38bdf8';
    ctx.lineWidth = 4;
    ctx.stroke();

    // Katana / tech element
    ctx.strokeStyle = '#94a3b8';
    ctx.lineWidth = 6;
    ctx.beginPath();
    ctx.moveTo(180, 120);
    ctx.lineTo(420, 520);
    ctx.stroke();

    return c.toDataURL('image/jpeg', 0.9);
  }

  function createCyberpunkStyleCanvas() {
    const c = document.createElement('canvas');
    c.width = 600;
    c.height = 600;
    const ctx = c.getContext('2d');

    // Neon night palette
    const grad = ctx.createRadialGradient(300, 300, 50, 300, 300, 400);
    grad.addColorStop(0, '#ff007f');
    grad.addColorStop(0.5, '#7928ca');
    grad.addColorStop(1, '#00f0ff');
    ctx.fillStyle = grad;
    ctx.fillRect(0, 0, 600, 600);

    // City grid lines & glowing beams
    ctx.strokeStyle = 'rgba(0, 240, 255, 0.4)';
    ctx.lineWidth = 3;
    for (let i = 0; i < 600; i += 40) {
      ctx.beginPath();
      ctx.moveTo(i, 0);
      ctx.lineTo(i, 600);
      ctx.stroke();

      ctx.beginPath();
      ctx.moveTo(0, i);
      ctx.lineTo(600, i);
      ctx.stroke();
    }

    // Volumetric rain particles
    ctx.fillStyle = 'rgba(255, 255, 255, 0.7)';
    for (let i = 0; i < 150; i++) {
      const rx = Math.random() * 600;
      const ry = Math.random() * 600;
      ctx.fillRect(rx, ry, 2, 8);
    }

    return c.toDataURL('image/jpeg', 0.9);
  }

  function createRenaissanceSubjectCanvas() {
    const c = document.createElement('canvas');
    c.width = 600;
    c.height = 600;
    const ctx = c.getContext('2d');

    ctx.fillStyle = '#1e1b18';
    ctx.fillRect(0, 0, 600, 600);

    // Modern Portrait pose
    ctx.fillStyle = '#e5d5c5';
    ctx.beginPath();
    ctx.arc(300, 240, 80, 0, Math.PI * 2);
    ctx.fill();

    ctx.fillStyle = '#42372d';
    ctx.beginPath();
    ctx.ellipse(300, 460, 160, 140, 0, 0, Math.PI * 2);
    ctx.fill();

    return c.toDataURL('image/jpeg', 0.9);
  }

  function createRenaissanceStyleCanvas() {
    const c = document.createElement('canvas');
    c.width = 600;
    c.height = 600;
    const ctx = c.getContext('2d');

    // Chiaroscuro warm golden umber oil palette
    const grad = ctx.createLinearGradient(0, 0, 600, 600);
    grad.addColorStop(0, '#fef08a');
    grad.addColorStop(0.3, '#b45309');
    grad.addColorStop(0.7, '#451a03');
    grad.addColorStop(1, '#0a0502');
    ctx.fillStyle = grad;
    ctx.fillRect(0, 0, 600, 600);

    // Oil impasto texture simulation
    ctx.fillStyle = 'rgba(217, 119, 6, 0.15)';
    for (let i = 0; i < 200; i++) {
      ctx.beginPath();
      ctx.arc(Math.random() * 600, Math.random() * 600, Math.random() * 30, 0, Math.PI * 2);
      ctx.fill();
    }

    return c.toDataURL('image/jpeg', 0.9);
  }

  function createGhibliSubjectCanvas() {
    const c = document.createElement('canvas');
    c.width = 600;
    c.height = 600;
    const ctx = c.getContext('2d');

    ctx.fillStyle = '#e0f2fe';
    ctx.fillRect(0, 0, 600, 600);

    // Cozy cat / creature outline
    ctx.fillStyle = '#334155';
    ctx.beginPath();
    ctx.arc(300, 320, 100, 0, Math.PI * 2);
    ctx.fill();

    // Cat ears
    ctx.beginPath();
    ctx.moveTo(220, 260);
    ctx.lineTo(240, 190);
    ctx.lineTo(280, 240);
    ctx.fill();

    ctx.beginPath();
    ctx.moveTo(380, 260);
    ctx.lineTo(360, 190);
    ctx.lineTo(320, 240);
    ctx.fill();

    return c.toDataURL('image/jpeg', 0.9);
  }

  function createGhibliStyleCanvas() {
    const c = document.createElement('canvas');
    c.width = 600;
    c.height = 600;
    const ctx = c.getContext('2d');

    // Lush watercolor rolling green hills & blue sky
    const sky = ctx.createLinearGradient(0, 0, 0, 350);
    sky.addColorStop(0, '#38bdf8');
    sky.addColorStop(1, '#bae6fd');
    ctx.fillStyle = sky;
    ctx.fillRect(0, 0, 600, 350);

    // Fluffy cloud
    ctx.fillStyle = '#ffffff';
    ctx.beginPath();
    ctx.arc(200, 140, 50, 0, Math.PI * 2);
    ctx.arc(260, 130, 70, 0, Math.PI * 2);
    ctx.arc(330, 140, 50, 0, Math.PI * 2);
    ctx.fill();

    // Green hill
    const grass = ctx.createLinearGradient(0, 350, 0, 600);
    grass.addColorStop(0, '#86efac');
    grass.addColorStop(1, '#15803d');
    ctx.fillStyle = grass;
    ctx.beginPath();
    ctx.arc(300, 700, 420, 0, Math.PI * 2);
    ctx.fill();

    return c.toDataURL('image/jpeg', 0.9);
  }

  // Bind Demo Buttons
  demoCyberpunk.addEventListener('click', () => {
    state.imageA = {
      dataUrl: createCyberpunkSubjectCanvas(),
      mime: 'image/jpeg',
      name: 'demo_cyber_samurai.jpg',
      size: '120 KB',
    };
    state.imageB = {
      dataUrl: createCyberpunkStyleCanvas(),
      mime: 'image/jpeg',
      name: 'demo_neon_city.jpg',
      size: '180 KB',
    };
    renderPreview(dropEmptyA, dropPreviewA, previewImgA, metaA, state.imageA);
    renderPreview(dropEmptyB, dropPreviewB, previewImgB, metaB, state.imageB);
    selectPreset('cyberpunk');
    showToast('Cyberpunk demo images loaded!', 'success');
  });

  demoRenaissance.addEventListener('click', () => {
    state.imageA = {
      dataUrl: createRenaissanceSubjectCanvas(),
      mime: 'image/jpeg',
      name: 'demo_portrait.jpg',
      size: '115 KB',
    };
    state.imageB = {
      dataUrl: createRenaissanceStyleCanvas(),
      mime: 'image/jpeg',
      name: 'demo_oil_texture.jpg',
      size: '190 KB',
    };
    renderPreview(dropEmptyA, dropPreviewA, previewImgA, metaA, state.imageA);
    renderPreview(dropEmptyB, dropPreviewB, previewImgB, metaB, state.imageB);
    selectPreset('renaissance');
    showToast('Renaissance demo images loaded!', 'success');
  });

  demoGhibli.addEventListener('click', () => {
    state.imageA = {
      dataUrl: createGhibliSubjectCanvas(),
      mime: 'image/jpeg',
      name: 'demo_creature.jpg',
      size: '110 KB',
    };
    state.imageB = {
      dataUrl: createGhibliStyleCanvas(),
      mime: 'image/jpeg',
      name: 'demo_ghibli_watercolor.jpg',
      size: '175 KB',
    };
    renderPreview(dropEmptyA, dropPreviewA, previewImgA, metaA, state.imageA);
    renderPreview(dropEmptyB, dropPreviewB, previewImgB, metaB, state.imageB);
    selectPreset('ghibli');
    showToast('Studio Ghibli demo images loaded!', 'success');
  });

  if (demoCatHat) {
    demoCatHat.addEventListener('click', () => {
      const catPrompt = 'A cute cat wearing a hat';
      if (customPromptInput) customPromptInput.value = catPrompt;
      state.prompt = catPrompt;
      showToast('Generating demo: "A cute cat wearing a hat"...', 'info');
      generateFromPrompt(null, catPrompt);
    });
  }

  // --------------------------------------------------------------------------
  // Preset Selection
  // --------------------------------------------------------------------------
  function selectPreset(presetId) {
    state.preset = presetId;
    document.querySelectorAll('.preset-card').forEach((card) => {
      if (card.dataset.preset === presetId) {
        card.classList.add('active');
      } else {
        card.classList.remove('active');
      }
    });
  }

  presetsCarousel.addEventListener('click', (e) => {
    const card = e.target.closest('.preset-card');
    if (card && card.dataset.preset) {
      selectPreset(card.dataset.preset);
    }
  });

  // --------------------------------------------------------------------------
  // Controls & Fine Tuning
  // --------------------------------------------------------------------------
  intensitySlider.addEventListener('input', (e) => {
    const val = parseInt(e.target.value, 10);
    state.intensity = val;
    let label = `${val}% (Harmonized)`;
    if (val < 40) label = `${val}% (Subtle)`;
    else if (val > 80) label = `${val}% (Radical)`;
    intensityVal.textContent = label;
  });

  // Quick Chips
  document.querySelectorAll('.chip').forEach((chip) => {
    chip.addEventListener('click', () => {
      const text = chip.textContent.replace(/^[^\w\s]+\s*/, '');
      if (customPromptInput.value) {
        customPromptInput.value += `, ${text}`;
      } else {
        customPromptInput.value = text;
      }
      state.prompt = customPromptInput.value;
    });
  });

  customPromptInput.addEventListener('input', (e) => {
    state.prompt = e.target.value;
  });

  // Reset All
  btnResetAll.addEventListener('click', () => {
    removeImage(true);
    removeImage(false);
    customPromptInput.value = '';
    state.prompt = '';
    intensitySlider.value = 65;
    intensityVal.textContent = '65% (Harmonized)';
    state.intensity = 65;
    selectPreset('auto');
    resetCanvasView();
    showToast('Studio workspace reset', 'info');
  });

  function resetCanvasView() {
    canvasEmpty.classList.remove('hidden');
    canvasLoading.classList.add('hidden');
    canvasResult.classList.add('hidden');
    viewModeToggle.classList.add('hidden');
    btnDownload.classList.add('hidden');
    btnCopyImg.classList.add('hidden');
  }

  // --------------------------------------------------------------------------
  // Fusion Synthesis API Invocation
  // --------------------------------------------------------------------------
  btnGenerate.addEventListener('click', startFusion);

  if (btnPromptGenerate) {
    btnPromptGenerate.addEventListener('click', () => {
      generateFromPrompt();
    });
  }

  if (btnAiGenSubject) {
    btnAiGenSubject.addEventListener('click', (e) => {
      e.stopPropagation();
      generateFromPrompt('A');
    });
  }

  // Shortcut Ctrl + Enter to synthesize
  window.addEventListener('keydown', (e) => {
    if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
      if (!state.isGenerating) startFusion();
    }
  });

  async function startFusion() {
    if (!state.imageA) {
      showToast('Please upload a Subject Image (Slot 01) first', 'error');
      dropAreaA.focus();
      return;
    }

    if (state.credits <= 0) {
      showToast('Out of free credits! Upgrade to Pro for unlimited fusions.', 'error');
      openPricingModal();
      return;
    }

    state.isGenerating = true;
    btnGenerate.disabled = true;
    showLoadingView();

    // Start progress timer
    const startTime = Date.now();
    let elapsedTimer = setInterval(() => {
      const elapsedSec = ((Date.now() - startTime) / 1000).toFixed(1);
      if (elapsedCounter) elapsedCounter.textContent = `Elapsed: ${elapsedSec}s`;
    }, 100);

    // Multi-stage indicator transitions
    if (progStep1) progStep1.classList.add('active');
    if (progStep2) progStep2.classList.remove('active');
    if (progStep3) progStep3.classList.remove('active');

    const stepTimeout1 = setTimeout(() => {
      if (loadingStageTitle) loadingStageTitle.textContent = 'Extracting Style Palette & Lighting...';
      if (loadingStageDesc) loadingStageDesc.textContent = 'Gemini Multimodal Vision is reading specular highlights and color temperature...';
      if (progStep2) progStep2.classList.add('active');
    }, 1800);

    const stepTimeout2 = setTimeout(() => {
      if (loadingStageTitle) loadingStageTitle.textContent = 'Denoising & Synthesizing Composition...';
      if (loadingStageDesc) loadingStageDesc.textContent = 'Harmonizing textures, shadows, and perspective...';
      if (progStep3) progStep3.classList.add('active');
    }, 3800);

    // 25-second watchdog to prevent infinite loading under network hangs
    const watchdogTimer = setTimeout(() => {
      if (state.isGenerating) {
        state.isGenerating = false;
        btnGenerate.disabled = false;
        clearInterval(elapsedTimer);
        clearTimeout(stepTimeout1);
        clearTimeout(stepTimeout2);
        showToast('Request took too long. Resetting canvas.', 'error');
        renderErrorState('Network timeout while communicating with Gemini API.');
      }
    }, 25000);

    try {
      const payload = {
        image_a_data: state.imageA.dataUrl,
        image_a_mime: state.imageA.mime,
        image_b_data: state.imageB ? state.imageB.dataUrl : '',
        image_b_mime: state.imageB ? state.imageB.mime : '',
        preset: state.preset,
        prompt: state.prompt,
        intensity: state.intensity,
      };

      const resp = await fetch('/api/fusion', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });

      const data = await resp.json();

      clearInterval(elapsedTimer);
      clearTimeout(stepTimeout1);
      clearTimeout(stepTimeout2);
      clearTimeout(watchdogTimer);

      if (!resp.ok || (data.error && !data.image_result)) {
        throw new Error(data.error || 'Server error occurred during fusion');
      } else {
        // Successful generation or synthesized harmonization!
        state.credits = Math.max(0, state.credits - 1);
        creditCount.textContent = state.credits;
        renderSuccessfulResult(data, false);
        showToast('Fusion artwork synthesized successfully!', 'success');
      }
    } catch (err) {
      clearInterval(elapsedTimer);
      clearTimeout(stepTimeout1);
      clearTimeout(stepTimeout2);
      clearTimeout(watchdogTimer);
      console.error('Fusion error:', err);
      showToast(err.message, 'error');
      renderErrorState(err.message);
    } finally {
      state.isGenerating = false;
      btnGenerate.disabled = false;
    }
  }

  function showLoadingView() {
    canvasEmpty.classList.add('hidden');
    canvasResult.classList.add('hidden');
    canvasLoading.classList.remove('hidden');
    viewModeToggle.classList.add('hidden');
    btnDownload.classList.add('hidden');
    btnCopyImg.classList.add('hidden');

    if (loadingStageTitle) loadingStageTitle.textContent = 'Synthesizing Creative Style Fusion...';
    if (loadingStageDesc) loadingStageDesc.textContent = 'Gemini Multimodal Vision is analyzing subject contours and style palette...';
    if (elapsedCounter) elapsedCounter.textContent = 'Elapsed: 0s';
  }

  function renderSuccessfulResult(data, hasNotice) {
    state.lastResult = data;
    canvasLoading.classList.add('hidden');
    canvasEmpty.classList.add('hidden');
    canvasResult.classList.remove('hidden');

    // Show toolbar
    viewModeToggle.classList.remove('hidden');
    btnDownload.classList.remove('hidden');
    btnCopyImg.classList.remove('hidden');

    // Output Image Result
    const finalImage = data.image_result || state.imageA.dataUrl;

    // Split Slider setup
    if (splitHandle) splitHandle.classList.remove('hidden');
    splitBaseImg.src = state.imageA.dataUrl;
    splitFusedImg.src = finalImage;
    updateSplitPosition(50);

    // Side by Side Setup
    sideImgSubject.src = state.imageA.dataUrl;
    sideImgStyle.src = state.imageB ? state.imageB.dataUrl : state.imageA.dataUrl;
    sideImgFused.src = finalImage;

    // AI Critique Card
    critiqueBody.textContent = data.analysis || 'Artistic harmonization completed seamlessly.';
    const duration = data.duration_ms ? `${(data.duration_ms / 1000).toFixed(1)}s` : '3.8s';
    critiqueMeta.textContent = `Preset: ${data.preset || state.preset} • Synthesized in ${duration}`;

    // Error or Notice Banner
    if (data.notice) {
      errorBanner.textContent = `ℹ️ ${data.notice}`;
      errorBanner.className = 'error-banner notice-banner';
      errorBanner.classList.remove('hidden');
    } else if (hasNotice && data.error) {
      errorBanner.textContent = `Notice: ${data.error}`;
      errorBanner.className = 'error-banner';
      errorBanner.classList.remove('hidden');
    } else {
      errorBanner.classList.add('hidden');
    }
  }

  function renderErrorState(errMsg) {
    canvasLoading.classList.add('hidden');
    canvasResult.classList.remove('hidden');
    errorBanner.textContent = `Image Generation Failed: ${errMsg}`;
    errorBanner.classList.remove('hidden');
    critiqueBody.textContent = 'Please verify that your local 9router is running on port 20128 and has a valid token.';
  }

  async function generateFromPrompt(targetSlot = null, overridePrompt = null) {
    let promptText = overridePrompt || state.prompt || (customPromptInput ? customPromptInput.value.trim() : '');
    if (!promptText) {
      promptText = window.prompt('Enter an image prompt for AI generation:', 'A cute cat wearing a hat');
      if (!promptText || !promptText.trim()) return;
      promptText = promptText.trim();
      if (customPromptInput) customPromptInput.value = promptText;
      state.prompt = promptText;
    }

    state.isGenerating = true;
    if (btnGenerate) btnGenerate.disabled = true;
    if (btnPromptGenerate) btnPromptGenerate.disabled = true;
    showLoadingView();

    if (loadingStageTitle) loadingStageTitle.textContent = 'Generating Image with AI...';
    if (loadingStageDesc) loadingStageDesc.textContent = `Routing to ag/gemini-3.1-flash-image: "${promptText}"`;

    const startTime = Date.now();
    let elapsedTimer = setInterval(() => {
      const elapsedSec = ((Date.now() - startTime) / 1000).toFixed(1);
      if (elapsedCounter) elapsedCounter.textContent = `Elapsed: ${elapsedSec}s`;
    }, 100);

    try {
      const resp = await fetch('/api/generate-image', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ prompt: promptText }),
      });

      clearInterval(elapsedTimer);

      const data = await resp.json();
      if (!resp.ok || data.error) {
        throw new Error(data.error || 'Failed to generate image');
      }

      const imgURI = data.image_result;

      if (targetSlot === 'A') {
        state.imageA = {
          dataUrl: imgURI,
          mime: 'image/png',
          name: 'ai_generated_subject.png',
          size: 'AI Generated',
        };
        renderPreview(dropEmptyA, dropPreviewA, previewImgA, metaA, state.imageA);
        showToast('Generated image set as Subject (Slot 01)!', 'success');
        resetCanvasView();
      } else {
        renderStandaloneImageResult(imgURI, promptText, data.duration_ms);
        showToast('AI Image generated successfully!', 'success');
      }
    } catch (err) {
      clearInterval(elapsedTimer);
      showToast(err.message, 'error');
      renderErrorState(err.message);
    } finally {
      state.isGenerating = false;
      if (btnGenerate) btnGenerate.disabled = false;
      if (btnPromptGenerate) btnPromptGenerate.disabled = false;
    }
  }

  function renderStandaloneImageResult(imgURI, promptText, durationMs) {
    const data = {
      image_result: imgURI,
      analysis: `Generated high-detail image using ag/gemini-3.1-flash-image for prompt: "${promptText}".`,
      preset: 'Text-to-Image',
      duration_ms: durationMs,
    };
    state.lastResult = data;
    canvasLoading.classList.add('hidden');
    canvasEmpty.classList.add('hidden');
    canvasResult.classList.remove('hidden');

    btnDownload.classList.remove('hidden');
    btnCopyImg.classList.remove('hidden');

    if (state.imageA && state.imageA.dataUrl) {
      viewModeToggle.classList.remove('hidden');
      if (splitHandle) splitHandle.classList.remove('hidden');
      splitBaseImg.src = state.imageA.dataUrl;
      splitFusedImg.src = imgURI;
      updateSplitPosition(50);
      sideImgSubject.src = state.imageA.dataUrl;
      sideImgStyle.src = state.imageB ? state.imageB.dataUrl : imgURI;
      sideImgFused.src = imgURI;
    } else {
      viewModeToggle.classList.add('hidden');
      switchViewMode('split');
      if (splitHandle) splitHandle.classList.add('hidden');
      splitBaseImg.src = imgURI;
      splitFusedImg.src = imgURI;
      updateSplitPosition(100);
      sideImgSubject.src = imgURI;
      sideImgStyle.src = imgURI;
      sideImgFused.src = imgURI;
    }

    critiqueBody.textContent = data.analysis;
    const durSec = durationMs ? `${(durationMs / 1000).toFixed(1)}s` : '2.5s';
    critiqueMeta.textContent = `Engine: ag/gemini-3.1-flash-image • Generated in ${durSec}`;
    errorBanner.classList.add('hidden');
  }

  // --------------------------------------------------------------------------
  // Interactive Split Comparison Slider (Before / After)
  // --------------------------------------------------------------------------
  function updateSplitPosition(pct) {
    state.splitPct = Math.max(0, Math.min(100, pct));
    const pos = `${state.splitPct}%`;
    // Wrapper stays full-size so the fused image never moves/rescales;
    // only the clip mask reveals more or less of it.
    splitFusedWrap.style.width = '100%';
    splitContainer.style.setProperty('--split-pos', pos);
    splitFusedWrap.style.clipPath = `polygon(0 0, ${pos} 0, ${pos} 100%, 0 100%)`;
    splitFusedWrap.style.webkitClipPath = `polygon(0 0, ${pos} 0, ${pos} 100%, 0 100%)`;
    splitHandle.style.left = pos;
  }

  let isDraggingSplit = false;

  function onSplitMove(clientX) {
    const rect = splitContainer.getBoundingClientRect();
    if (!rect.width) return;
    const offsetX = clientX - rect.left;
    const pct = (offsetX / rect.width) * 100;
    updateSplitPosition(pct);
  }

  splitContainer.addEventListener('mousedown', (e) => {
    e.preventDefault();
    isDraggingSplit = true;
    onSplitMove(e.clientX);
  });

  window.addEventListener('mousemove', (e) => {
    if (!isDraggingSplit) return;
    onSplitMove(e.clientX);
  });

  window.addEventListener('mouseup', () => {
    isDraggingSplit = false;
  });

  // Touch Support for mobile / tablets
  splitContainer.addEventListener('touchstart', (e) => {
    isDraggingSplit = true;
    if (e.touches && e.touches[0]) onSplitMove(e.touches[0].clientX);
  }, { passive: true });

  window.addEventListener('touchmove', (e) => {
    if (!isDraggingSplit) return;
    if (e.touches && e.touches[0]) onSplitMove(e.touches[0].clientX);
  }, { passive: true });

  window.addEventListener('touchend', () => {
    isDraggingSplit = false;
  });

  // View Mode Toggle (Split vs. Side)
  function switchViewMode(mode) {
    if (mode === 'side') {
      if (modeSide) modeSide.classList.add('active');
      if (modeSplit) modeSplit.classList.remove('active');
      if (splitContainer) splitContainer.classList.add('hidden');
      if (sideContainer) sideContainer.classList.remove('hidden');
      state.viewMode = 'side';
    } else {
      if (modeSplit) modeSplit.classList.add('active');
      if (modeSide) modeSide.classList.remove('active');
      if (splitContainer) splitContainer.classList.remove('hidden');
      if (sideContainer) sideContainer.classList.add('hidden');
      state.viewMode = 'split';
    }
  }

  if (modeSplit) {
    modeSplit.addEventListener('click', () => switchViewMode('split'));
  }

  if (modeSide) {
    modeSide.addEventListener('click', () => switchViewMode('side'));
  }

  // --------------------------------------------------------------------------
  // Download & Copy Artwork
  // --------------------------------------------------------------------------
  btnDownload.addEventListener('click', () => {
    if (!state.lastResult || !state.lastResult.image_result) return;
    const a = document.createElement('a');
    a.href = state.lastResult.image_result;
    a.download = `chromafuse-${state.preset}-${Date.now()}.png`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    showToast('Artwork download started!', 'success');
  });

  btnCopyImg.addEventListener('click', async () => {
    if (!state.lastResult || !state.lastResult.image_result) return;
    try {
      // Fetch blob from data URI
      const res = await fetch(state.lastResult.image_result);
      const blob = await res.blob();
      await navigator.clipboard.write([
        new ClipboardItem({ [blob.type]: blob }),
      ]);
      showToast('Artwork copied to clipboard!', 'success');
    } catch {
      showToast('Clipboard copy not supported by your browser', 'error');
    }
  });

  // --------------------------------------------------------------------------
  // AI Creative Director Chat Drawer
  // --------------------------------------------------------------------------
  function toggleDrawer() {
    assistantDrawer.classList.toggle('hidden');
  }

  btnToggleAssistant.addEventListener('click', toggleDrawer);
  btnCloseDrawer.addEventListener('click', toggleDrawer);

  chatForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    const text = userInput.value.trim();
    if (!text) return;

    userInput.value = '';
    appendChatMessage('user', text);

    try {
      const resp = await fetch('/api/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          message: text,
          persona: 'creative_director',
          history: state.chatHistory,
        }),
      });

      const data = await resp.json();
      if (!resp.ok || data.error) {
        throw new Error(data.error || 'Failed to chat with AI Director');
      }

      state.chatHistory = data.history || [];
      appendChatMessage('bot', data.reply);
    } catch (err) {
      appendChatMessage('bot', `⚠️ *Error:* ${err.message}`);
    }
  });

  function appendChatMessage(role, text) {
    const msg = document.createElement('div');
    msg.className = `message ${role === 'user' ? 'user-message' : 'bot-message'}`;

    let parsedHtml = text;
    if (window.marked && window.DOMPurify) {
      parsedHtml = DOMPurify.sanitize(marked.parse(text));
    }

    msg.innerHTML = `
      <div class="avatar">${role === 'user' ? '👤' : '🎨'}</div>
      <div class="message-body">
        <div class="bubble">${parsedHtml}</div>
        <div class="message-meta"><span class="message-time">Just now</span></div>
      </div>
    `;

    chatMessages.appendChild(msg);
    chatMessages.scrollTop = chatMessages.scrollHeight;
  }

  // --------------------------------------------------------------------------
  // Monetization Modal
  // --------------------------------------------------------------------------
  function openPricingModal() {
    pricingModal.classList.remove('hidden');
  }

  function closePricingModal() {
    pricingModal.classList.add('hidden');
  }

  btnOpenPricing.addEventListener('click', openPricingModal);
  btnCloseModal.addEventListener('click', closePricingModal);
  pricingModal.addEventListener('click', (e) => {
    if (e.target === pricingModal) closePricingModal();
  });

  btnSimCheckout.addEventListener('click', () => {
    closePricingModal();
    state.credits += 250;
    creditCount.textContent = state.credits;
    showToast('🎉 Upgrade Successful! 250 Pro Studio Credits added.', 'success');
  });

  btnSimStudio.addEventListener('click', () => {
    closePricingModal();
    showToast('Studio Enterprise demo inquiry recorded!', 'info');
  });

  // --------------------------------------------------------------------------
  // Initialization
  // --------------------------------------------------------------------------
  async function loadPresetsFromAPI() {
    try {
      const resp = await fetch('/api/presets');
      if (resp.ok) {
        const presets = await resp.json();
        if (Array.isArray(presets) && presets.length > 0) {
          presetsCarousel.innerHTML = '';
          presets.forEach((p, idx) => {
            const card = document.createElement('div');
            const isActive = (state.preset && p.id === state.preset) || (!state.preset && idx === 0);
            card.className = `preset-card ${isActive ? 'active' : ''}`;
            card.dataset.preset = p.id;
            card.innerHTML = `
              <div class="preset-emoji">${p.emoji || '✨'}</div>
              <div class="preset-info">
                <div class="preset-top">
                  <span class="preset-title">${p.name}</span>
                  <span class="preset-tag">${p.badge || 'Art'}</span>
                </div>
                <p class="preset-desc">${p.description}</p>
              </div>
            `;
            presetsCarousel.appendChild(card);
          });
        }
      }
    } catch {
      // Keep static DOM fallback presets
    }
  }

  loadPresetsFromAPI();
})();
