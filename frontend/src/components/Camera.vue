<template>
  <div class="monitor-window">
    <div class="main-container">
      <section class="viewport-section">
        <div class="video-canvas">
          <img ref="streamImgRef" class="stream-img" src="" style="display: none;" />

          <div v-if="!isOnline" class="signal-lost">
            <div class="glitch-text">NO SIGNAL</div>
            <p>等待视频流推送到终端...</p>
          </div>

          <div class="osd-overlay">
            <div class="rec-indicator">
              <span class="rec-dot"></span>
              <span class="rec-text">REC</span>
            </div>
          </div>

          <div class="corner top-left"></div>
          <div class="corner top-right"></div>
          <div class="corner bottom-left"></div>
          <div class="corner bottom-right"></div>
        </div>
      </section>

      <aside class="params-sidebar">
        <div class="sidebar-title">
          <i class="ri-equalizer-fill"></i>
          <span>相机控制参数</span>
        </div>

        <div class="params-list custom-scrollbar">
          <!-- 1. 曝光 -->
          <div class="param-item">
            <div class="param-header">
              <label>曝光 (Exposure)</label>
              <select v-model="cameraParams.exposure.mode" class="mode-select" @change="applyExposureAuto">
                <option value="manual">手动</option>
                <option value="once_auto">单次自动</option>
                <option value="continuous_auto">连续自动</option>
              </select>
            </div>
            <div v-if="cameraParams.exposure.mode === 'manual'" class="param-controls">
              <div class="slider-input-row">
                <input 
                  type="range" 
                  v-model.number="cameraParams.exposure.value" 
                  min="1" 
                  max="100000" 
                  step="1"
                  class="industrial-slider"
                  @change="applyExposureTime"
                />
                <input 
                  type="number" 
                  v-model.number="cameraParams.exposure.value" 
                  min="1" 
                  max="100000" 
                  step="1"
                  class="param-input"
                  @keyup.enter="applyExposureTime"
                />
                <span class="param-unit">ms</span>
              </div>
              
            </div>
          </div>

          <!-- 2. 增益 -->
          <div class="param-item">
            <div class="param-header">
              <label>增益 (Gain)</label>
              <select v-model="cameraParams.gain.mode" class="mode-select" @change="applyGainAuto">
                <option value="manual">手动</option>
                <option value="once_auto">单次自动</option>
                <option value="continuous_auto">连续自动</option>
              </select>
            </div>
            <div v-if="cameraParams.gain.mode === 'manual'" class="param-controls">
              <div class="slider-input-row">
                <input 
                  type="range" 
                  v-model.number="cameraParams.gain.value" 
                  min="0" 
                  max="24" 
                  step="0.1"
                  class="industrial-slider"
                  @change="applyGain"
                />
                <input 
                  type="number" 
                  v-model.number="cameraParams.gain.value" 
                  min="0" 
                  max="24" 
                  step="0.1"
                  class="param-input"
                  @keyup.enter="applyGain"
                />
                <span class="param-unit">dB</span>
              </div>
            </div>
          </div>

          <!-- 3. 数字增益 -->
          <div class="param-item">
            <div class="param-header">
              <label>数字增益 (Digital Gain)</label>
              <select v-model="cameraParams.digitalGain.mode" class="mode-select" @change="applyDigitalGainMode">
                <option value="manual">手动</option>
                <option value="once_auto">单次自动</option>
                <option value="continuous_auto">连续自动</option>
              </select>
            </div>
            <div v-if="cameraParams.digitalGain.mode === 'manual'" class="param-controls">
              <div class="slider-input-row">
                <input 
                  type="range" 
                  v-model.number="cameraParams.digitalGain.value" 
                  min="1" 
                  max="16" 
                  step="0.1"
                  class="industrial-slider"
                  @change="applyDigitalGain"
                />
                <input 
                  type="number" 
                  v-model.number="cameraParams.digitalGain.value" 
                  min="1" 
                  max="16" 
                  step="0.1"
                  class="param-input"
                  @keyup.enter="applyDigitalGain"
                />
                <span class="param-unit">x</span>
              </div>
            </div>
          </div>

          <!-- 4. 白平衡 -->
          <!-- <div class="param-item">
            <div class="param-header">
              <label>白平衡 (White Balance)</label>
              <select v-model="cameraParams.whiteBalance.mode" class="mode-select" @change="applyBalanceWhiteAuto">
                <option value="manual">手动</option>
                <option value="once_auto">单次自动</option>
                <option value="continuous_auto">连续自动</option>
              </select>
            </div>
            <div v-if="cameraParams.whiteBalance.mode === 'manual'" class="param-controls">
              <div class="slider-input-row">
                <input 
                  type="range" 
                  v-model.number="cameraParams.whiteBalance.value" 
                  min="2000" 
                  max="10000" 
                  step="100"
                  class="industrial-slider"
                  @change="applyWhiteBalance"
                />
                <input 
                  type="number" 
                  v-model.number="cameraParams.whiteBalance.value" 
                  min="2000" 
                  max="10000" 
                  step="100"
                  class="param-input"
                  @keyup.enter="applyWhiteBalance"
                />
                <span class="param-unit">Ratio</span>
              </div>
            </div>
          </div> -->

          <!-- 5. 伽马校正 -->
          <div class="param-item">
            <div class="param-header">
              <label>伽马校正 (Gamma)</label>
              <select v-model="cameraParams.gamma.enabled" class="mode-select" @change="applyGammaEnabled">
                <option :value="false">关闭</option>
                <option :value="true">开启</option>
              </select>
            </div>
            <div v-if="cameraParams.gamma.enabled" class="param-controls">
              <div class="slider-input-row">
                <input 
                  type="range" 
                  v-model.number="cameraParams.gamma.value" 
                  min="0.1" 
                  max="3.0" 
                  step="0.1"
                  class="industrial-slider"
                  @change="applyGamma"
                />
                <input 
                  type="number" 
                  v-model.number="cameraParams.gamma.value" 
                  min="0.1" 
                  max="3.0" 
                  step="0.1"
                  class="param-input"
                  @keyup.enter="applyGamma"
                />
              </div>
            </div>
          </div>

          <!-- 6. 水平翻转 -->
          <div class="param-item">
            <div class="toggle-row">
              <label>水平翻转 (Flip H)</label>
              <button 
                :class="['toggle-button', { active: cameraParams.flipHorizontal }]"
                @click="cameraParams.flipHorizontal = !cameraParams.flipHorizontal"
              >
                <span class="toggle-inner"></span>
              </button>
            </div>
          </div>

          <!-- 7. 垂直翻转 -->
          <div class="param-item">
            <div class="toggle-row">
              <label>垂直翻转 (Flip V)</label>
              <button 
                :class="['toggle-button', { active: cameraParams.flipVertical }]"
                @click="cameraParams.flipVertical = !cameraParams.flipVertical"
              >
                <span class="toggle-inner"></span>
              </button>
            </div>
          </div>
        </div>

        <div class="sidebar-footer">
          <div class="connection-status">
            <span class="status-dot" :class="{ 'online': isOnline }"></span>
            {{ isOnline ? '视频流已连接' : '视频流离线' }}
          </div>
        </div>
      </aside>
    </div>
    
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted, computed } from 'vue'
import { Events } from '@wailsio/runtime'
import { 
  GetCameraParams,

  SetExposureAuto,
  SetExposureTime,
  SetGainAuto,
  SetGain,
  SetDigitalGainEnabled,
  SetDigitalGain,
  SetBalanceWhiteAuto,
  SetWhiteBalanceRed,
  SetGammaEnabled,
  SetGamma,
} from "../../bindings/changeme/backend/hikcameraservice";

const streamImgRef = ref(null)
const isOnline = ref(false)
const currentTime = ref(new Date())
let offCameraFrame

// 参数模式选项 - 使用 reactive 统一管理
const cameraParams = reactive({
  exposure: {
    mode: 'manual', // manual, once_auto, continuous_auto
    value: 20.0
  },
  gain: {
    mode: 'manual',
    value: 0.0
  },
  digitalGain: {
    mode: 'manual',
    value: 1.0
  },
  whiteBalance: {
    mode: 'manual',
    value: 3200
  },
  gamma: {
    enabled: false,
    value: 1.0
  },
  flipHorizontal: false,
  flipVertical: false
})

let timeTicker

// 刷新顶部装饰条系统时间
const updateTime = () => {
  currentTime.value = new Date()
}

const currentFullTime = computed(() => {
  const now = currentTime.value
  const date = now.toLocaleDateString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' }).replace(/\//g, '-')
  const time = now.toLocaleTimeString('zh-CN', { hour12: false })
  return `${date} ${time}`
})

const enumModeToFrontend = (value) => {
  switch (value) {
    case 1:
      return 'once_auto'
    case 2:
      return 'continuous_auto'
    default:
      return 'manual'
  }
}

const showCameraFrame = (payload) => {
  const data = payload?.data ?? payload
  if (!streamImgRef.value || !data?.image) {
    return
  }

  streamImgRef.value.src = data.image
  if (!isOnline.value) {
    isOnline.value = true
    streamImgRef.value.style.display = 'block'
  }
}

const loadCameraParams = async () => {
  const params = await GetCameraParams()

  cameraParams.exposure.mode = enumModeToFrontend(params.exposureAuto)
  cameraParams.exposure.value = params.exposureTime
  cameraParams.gain.mode = enumModeToFrontend(params.gainAuto)
  cameraParams.gain.value = params.gain
  cameraParams.digitalGain.mode = params.digitalGainEnable ? 'manual' : 'continuous_auto'
  cameraParams.digitalGain.value = params.digitalGain
  cameraParams.whiteBalance.mode = enumModeToFrontend(params.balanceWhiteAuto)
  cameraParams.whiteBalance.value = params.balanceWhite
  cameraParams.gamma.enabled = params.gammaEnable
  cameraParams.gamma.value = params.gamma
}

onMounted(async () => {
  timeTicker = setInterval(updateTime, 1000)
  offCameraFrame = Events.On('hik_camera_frame', showCameraFrame)
  await loadCameraParams()
})


onUnmounted(() => {
  clearInterval(timeTicker)
  offCameraFrame?.()
})

const applyExposureAuto = async () => {
  try {
    await SetExposureAuto(cameraParams.exposure.mode)
    if (cameraParams.exposure.mode === 'manual') {
      await SetExposureTime(cameraParams.exposure.value)
    }
  } catch (err) {
    console.error(err)
  }
}

const applyExposureTime = async () => {
  if (cameraParams.exposure.mode !== 'manual') {
    return
  }
  try {
    await SetExposureTime(cameraParams.exposure.value)
  } catch (err) {
    console.error(err)
  }
}

const applyGainAuto = async () => {
  try {
    await SetGainAuto(cameraParams.gain.mode)
    if (cameraParams.gain.mode === 'manual') {
      await SetGain(cameraParams.gain.value)
    }
  } catch (err) {
    console.error(err)
  }
}

const applyGain = async () => {
  if (cameraParams.gain.mode !== 'manual') {
    return
  }
  try {
    await SetGain(cameraParams.gain.value)
  } catch (err) {
    console.error(err)
  }
}

const applyDigitalGainMode = async () => {
  try {
    const enabled = cameraParams.digitalGain.mode === 'manual'
    await SetDigitalGainEnabled(enabled)
    if (enabled) {
      await SetDigitalGain(cameraParams.digitalGain.value)
    }
  } catch (err) {
    console.error(err)
  }
}

const applyDigitalGain = async () => {
  if (cameraParams.digitalGain.mode !== 'manual') {
    return
  }
  try {
    await SetDigitalGain(cameraParams.digitalGain.value)
  } catch (err) {
    console.error(err)
  }
}

const applyBalanceWhiteAuto = async () => {
  try {
    await SetBalanceWhiteAuto(cameraParams.whiteBalance.mode)
    if (cameraParams.whiteBalance.mode === 'manual') {
      await SetWhiteBalanceRed(cameraParams.whiteBalance.value)
    }
  } catch (err) {
    console.error(err)
  }
}

const applyWhiteBalance = async () => {
  if (cameraParams.whiteBalance.mode !== 'manual') {
    return
  }
  try {
    await SetWhiteBalanceRed(cameraParams.whiteBalance.value)
  } catch (err) {
    console.error(err)
  }
}

const applyGammaEnabled = async () => {
  try {
    await SetGammaEnabled(cameraParams.gamma.enabled)
    if (cameraParams.gamma.enabled) {
      await SetGamma(cameraParams.gamma.value)
    }
  } catch (err) {
    console.error(err)
  }
}

const applyGamma = async () => {
  if (!cameraParams.gamma.enabled) {
    return
  }
  try {
    await SetGamma(cameraParams.gamma.value)
  } catch (err) {
    console.error(err)
  }
}


</script>

<style scoped>

.monitor-window {


  display: flex;
  position: fixed;
  inset: 0;
  background: #1e293b;
  color: #cbd5e1;
  font-family: 'Inter', 'Segoe UI', system-ui, sans-serif;
  overflow: hidden;
}

/* 顶部时间条样式（若主视图不需要，可保持在 template 中注释） */
.monitor-header {
  height: 40px;
  background: #0f172a;
  border-bottom: 1px solid #1e293b;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 16px;
  font-size: 13px;
  flex-shrink: 0;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #3b82f6;
  font-weight: bold;
}

.terminal-id {
  color: #475569;
  font-size: 11px;
}

/* 主容器 */
.main-container {
  flex: 1;
  display: flex;
  overflow: hidden;
  min-width: 0;
}

/* 视频预览区 */
.viewport-section {
  flex: 1;
  background: #000000;
  padding: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;
  min-width: 0;
}

.video-canvas {
  width: 100%;
  height: 100%;
  position: relative;
  background: #050505;
  border: 1px solid #1e293b;
  overflow: hidden; /* 防止高频大比例拉伸时画面溢出 */
}

/* 🛠️ 核心修复：移除 Flex 弹性自适应，改用标准绝对定位 100% 贴靠 */
.stream-img {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  /* 采用 contain，保证工业相机原始比例视场完整，两侧自适应对称黑边，绝不产生向右下角的剪切漂移 */
  object-fit: contain; 
  object-position: center;
  z-index: 1;
}

/* OSD 叠加层 */
.osd-overlay {
  position: absolute;
  top: 15px;
  right: 15px;
  pointer-events: none;
  z-index: 10;
}

.rec-indicator {
  display: flex;
  align-items: center;
  gap: 8px;
  background: rgba(0, 0, 0, 0.6);
  border: 1px solid rgba(239, 68, 68, 0.2);
  padding: 4px 10px;
  border-radius: 4px;
}

.rec-dot {
  width: 10px;
  height: 10px;
  background: #ef4444;
  border-radius: 50%;
  box-shadow: 0 0 10px #ef4444;
  animation: blink 1s steps(2, start) infinite;
}

.rec-text {
  color: #ef4444;
  font-weight: 900;
  font-size: 14px;
}

@keyframes blink { to { opacity: 0; } }

/* 信号丢失样式：基于绝对定位进行自适应屏幕几何居中 */
.signal-lost {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  text-align: center;
  color: #475569;
  z-index: 2;
}

.glitch-text {
  font-size: 36px;
  font-weight: 900;
  letter-spacing: 5px;
  color: #ef4444;
  margin-bottom: 10px;
}

/* 装饰边角 - 指示相控系统的科技感边界 */
.corner {
  position: absolute;
  width: 16px;
  height: 16px;
  border: 2px solid #3b82f6; /* 靓丽科技蓝 */
  pointer-events: none;
  z-index: 10;
}
.top-left { top: 6px; left: 6px; border-right: none; border-bottom: none; }
.top-right { top: 6px; right: 6px; border-left: none; border-bottom: none; }
.bottom-left { bottom: 6px; left: 6px; border-right: none; border-top: none; }
.bottom-right { bottom: 6px; right: 6px; border-left: none; border-top: none; }

/* 右侧侧边栏 */
.params-sidebar {
  width: 260px;
  background: #0f172a;
  border-left: 1px solid #1e293b;
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
}

.sidebar-title {
  padding: 16px;
  font-size: 14px;
  font-weight: bold;
  border-bottom: 1px solid #1e293b;
  display: flex;
  align-items: center;
  gap: 8px;
  background: #1e293b55;
}

.params-list {
  flex: 1;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 24px;
  overflow-y: auto;
}

.param-item {
  display: flex;
  flex-direction: column;
}

.param-info {
  display: flex;
  justify-content: space-between;
  margin-bottom: 10px;
}

.param-info label {
  font-size: 11px;
  color: #94a3b8;
  text-transform: uppercase;
}

/* 参数头部 */
.param-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.param-header label {
  font-size: 11px;
  color: #94a3b8;
  text-transform: uppercase;
}

/* 模式选择器 - 下拉框 */
.mode-select {
  background: #1e293b;
  border: 1px solid #334155;
  color: #f1f5f9;
  font-size: 11px;
  padding: 4px 8px;
  border-radius: 3px;
  cursor: pointer;
  min-width: 80px;
  outline: none;
  transition: border-color 0.2s ease;
}

.mode-select:hover {
  border-color: #475569;
}

.mode-select:focus {
  border-color: #3b82f6;
}

.mode-select option {
  background: #1e293b;
  color: #f1f5f9;
}

/* 参数控制区 */
.param-controls {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

/* 滑动条和输入框并排 */
.slider-input-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.slider-input-row .industrial-slider {
  flex: 1;
}

.param-input {
  width: 60px;
  background: #1e293b;
  border: 1px solid #334155;
  color: #f1f5f9;
  font-size: 12px;
  padding: 4px 6px;
  border-radius: 3px;
  text-align: center;
}

.param-input:focus {
  outline: none;
  border-color: #3b82f6;
}

/* 参数单位 */
.param-unit {
  font-size: 10px;
  color: #64748b;
  text-align: right;
}

/* 开关行 */
.toggle-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.toggle-row label {
  font-size: 12px;
  color: #94a3b8;
}

/* 开关按钮 */
.toggle-button {
  width: 44px;
  height: 24px;
  background: #334155;
  border: none;
  border-radius: 12px;
  cursor: pointer;
  position: relative;
  transition: background 0.2s ease;
  padding: 0;
}

.toggle-button.active {
  background: #3b82f6;
}

.toggle-inner {
  position: absolute;
  top: 2px;
  left: 2px;
  width: 20px;
  height: 20px;
  background: #f1f5f9;
  border-radius: 50%;
  transition: transform 0.2s ease;
}

.toggle-button.active .toggle-inner {
  transform: translateX(20px);
}

.param-value {
  color: #3b82f6;
  font-weight: bold;
  font-size: 12px;
}

/* 滑动条样式 */
.industrial-slider {
  width: 100%;
  accent-color: #3b82f6;
  cursor: pointer;
}

.param-presets {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 5px;
  margin-top: 8px;
}

.param-presets button {
  background: #1e293b;
  border: 1px solid #334155;
  color: #64748b;
  font-size: 10px;
  padding: 4px;
  border-radius: 3px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.param-presets button:hover {
  border-color: #3b82f6;
  color: #f1f5f9;
}

.disabled { opacity: 0.3; pointer-events: none; }

.sidebar-footer {
  padding: 16px;
  border-top: 1px solid #1e293b;
  font-size: 11px;
  color: #64748b;
}

.connection-status {
  display: flex;
  align-items: center;
  gap: 8px;
}

.status-dot {
  width: 6px;
  height: 6px;
  background: #475569;
  border-radius: 50%;
}

.status-dot.online {
  background: #10b981;
  box-shadow: 0 0 5px #10b981;
}

/* 极细微滚动条 */
.custom-scrollbar::-webkit-scrollbar { width: 4px; }
.custom-scrollbar::-webkit-scrollbar-thumb { background: #1e293b; border-radius: 10px; }
</style>