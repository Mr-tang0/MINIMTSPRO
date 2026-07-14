<template>
  <div class="project-settings-content">

    <div class="settings-body scrollable">
      <!-- 1. 基本信息 -->
      <div class="config-section">
        <div class="section-tag">基本信息</div>
        <div class="form-row">
          <div class="form-item flex-1">
            <label>实验人</label>
            <input v-model="form.experimenter" type="text" spellcheck="false" placeholder="请输入姓名" />
          </div>
          <div class="form-item flex-1">
            <label>样品编号</label>
            <input v-model="form.sampleNo" type="text" placeholder="请输入编号" />
          </div>
          <div class="form-item flex-1">
            <label>实验时间</label>
            <input v-model="form.testDate" type="text" readonly class="readonly" />
          </div>
        </div>
      </div>

      <!-- 2. 准静态测试参数 -->
      <div class="config-section">
        <div class="section-tag">准静态测试参数</div>
        <div class="quasi-grid">
          <!-- 形状选择 -->
          <div class="sub-pane border-r">
            <label class="pane-title">1. 样品形状</label>
            <select v-model="form.sampleShape" class="standard-select">
              <option value="dogbone">狗骨头状 (Dog-bone)</option>
              <option value="cylinder">圆柱状 (Cylinder)</option>
            </select>
            <div class="sample-preview">
              <svg v-if="form.sampleShape === 'dogbone'" viewBox="0 0 100 60" class="sample-svg">
                <path d="M10,20 L30,20 C35,20 35,28 40,28 L60,28 C65,28 65,20 70,20 L90,20 L90,40 L70,40 C65,40 65,32 60,32 L40,32 C35,32 35,40 30,40 L10,40 Z" fill="none" stroke="#3b82f6" stroke-width="1.5" />
                <line x1="40" y1="45" x2="60" y2="45" stroke="#64748b" stroke-width="1" stroke-dasharray="2,2" />
                <text x="50" y="55" text-anchor="middle" font-size="8" fill="#64748b">L0 标距段</text>
              </svg>
              <svg v-else viewBox="0 0 100 60" class="sample-svg">
                <ellipse cx="50" cy="15" rx="25" ry="8" fill="none" stroke="#3b82f6" stroke-width="1.5" />
                <path d="M25,15 L25,45 A25,8 0 0,0 75,45 L75,15" fill="none" stroke="#3b82f6" stroke-width="1.5" />
                <ellipse cx="50" cy="45" rx="25" ry="8" fill="none" stroke="#3b82f6" stroke-width="1.5" stroke-dasharray="2,2" />
                <line x1="82" y1="15" x2="82" y2="45" stroke="#64748b" stroke-width="1" stroke-dasharray="2,2" />
                <text x="88" y="32" text-anchor="start" font-size="8" fill="#64748b">L0</text>
              </svg>
            </div>
          </div>

          <!-- 尺寸输入 -->
          <div class="sub-pane border-r">
            <label class="pane-title">2. 尺寸参数 (mm)</label>
            <div class="dim-inputs" v-if="form.sampleShape === 'dogbone'">
              <div class="input-unit-group"><span>宽度 W</span><input v-model.number="form.width" type="number" step="0.01" /></div>
              <div class="input-unit-group"><span>厚度 T</span><input v-model.number="form.thickness" type="number" step="0.01" /></div>
              <div class="input-unit-group highlight"><span>标距 L0</span><input v-model.number="form.sectionLength" type="number" step="0.01" @input="autoCalcSpeed" /></div>
              <div class="area-calc">计算截面积: <span>{{ form.width }} × {{ form.thickness }} = {{ currentArea }} </span> mm²</div>
            </div>
            <div class="dim-inputs" v-else>
              <div class="input-unit-group"><span>直径 D</span><input v-model.number="form.diameter" type="number" step="0.01" /></div>
              <div class="input-unit-group highlight"><span>标距 L0</span><input v-model.number="form.sectionLength" type="number" step="0.01" @input="autoCalcSpeed" /></div>
              <div class="area-calc">计算截面积: <span>π × {{ form.diameter }}² = {{ currentArea }} </span> mm²</div>
            </div>
            
          </div>

          <!-- 测试控制 -->
          <div class="sub-pane">
            <label class="pane-title">3. 测试控制</label>
            <div class="form-item">
              <label>试验类型</label>
              <div class="toggle-group">
                <button :class="{ active: form.type === 'tension' }" @click="form.type = 'tension'">拉伸 (T)</button>
                <button :class="{ active: form.type === 'compression' }" @click="form.type = 'compression'">压缩 (C)</button>
              </div>
            </div>
            <div class="form-item mt-10">
              <label>运行速度 (mm/s)</label>
              <input v-model.number="form.speed" type="number" step="0.001" :class="{ 'danger-text': isSpeedOverridden }" @input="handleSpeedManualChange" />
            </div>
            <div class="form-item mt-10">
              <label>停止条件</label>
              <select v-model="form.stopCondition">
                <option value="manual">手动停止</option>
                <option value="time">按时间 (s)</option>
                <option value="load">按载荷 (N)</option>
                <option value="disp">按位移 (mm)</option>
              </select>
            </div>
          </div>
        </div>

        <!-- 保存与导出：三栏下方左右排列 -->
        <div class="quasi-footer border-t mt-15 pt-15">
          <div class="form-item flex-2">
            <label>准静态数据保存路径</label>
            <div class="path-picker">
              <input v-model="form.filePath" type="text" readonly />
              <button class="btn-picker" @click="handleSelectDir('filePath')"><i class="ri-folder-open-fill"></i></button>
            </div>
          </div>
          <div class="form-item flex-1">
            <label>保存文件名</label>
            <input v-model="form.fileName" type="text" placeholder="Result_001" />
          </div>
        </div>
      </div>

      <!-- 3. DIC 与 视频引伸 并排 -->
      <div v-if="props.isCameraConnected" class="dual-row">
        <!-- DIC 配置 -->
        <div class="config-section flex-1">
          <div class="section-tag color-dic">DIC 采集配置</div>
          <div class="split-layout">
            <div class="left-controls">
              <label class="check-item">
                <input type="checkbox" v-model="form.dicEnable" />
                <span class="mark"></span> 相机拍摄
              </label>
              <label class="check-item">
                <input type="checkbox" v-model="form.externalTrigger" />
                <span class="mark"></span> 发射触发信号
              </label>
            </div>
            <div class="right-display">
              <div v-if="!form.dicEnable && !form.externalTrigger" class="placeholder">
                <i class="ri-camera-lens-line"></i>
                <span>启用 DIC 配置</span>
              </div>
              
              <div v-if="form.dicEnable" class="expand-group animate-in">
                <div class="form-item">
                  <label>图片保存路径</label>
                  <div class="path-picker">
                    <input v-model="form.dicFolder" type="text" placeholder="选择保存目录" />
                    <button class="btn-picker" @click="handleSelectDir('dicFolder')"><i class="ri-image-add-fill"></i></button>
                  </div>
                </div>
                <div class="form-item mt-5">
                  <label>命名前缀</label>
                  <input v-model="form.dicFileName" type="text" />
                </div>
              </div>

              <div v-if="form.externalTrigger" class="expand-group border-t mt-10 pt-10 animate-in">
                <div class="form-row">
                  <div class="form-item flex-1">
                    <label>触发类型</label>
                    <select v-model="form.triggerType">
                      <option value="internal">固定频率</option>
                      <option value="external">外部同步</option>
                    </select>
                  </div>
                  <div class="form-item flex-1">
                    <label>间隔(ms)</label>
                    <input v-model="form.triggerInterval" type="number" />
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 视频引伸计 -->
        <div v-if="props.isCameraConnected" class="config-section flex-1">
          <div class="section-tag color-video">视频引伸计 (AVE)</div>
          <div class="split-layout">
            <div class="left-controls">
              <label class="check-item text-cyan">
                <input type="checkbox" v-model="form.videoExtEnable" />
                <span class="mark"></span> 变形测算
              </label>
              <label v-if="form.videoExtEnable" class="check-item poisson-check">
                <input type="checkbox" v-model="form.poissonEnable" />
                <span class="mark"></span> 横向变形
              </label>
            </div>
            <div class="right-display">
              <div v-if="!form.videoExtEnable" class="placeholder">
                <svg viewBox="0 0 100 60" class="diag-svg">
                   <rect x="42" y="10" width="16" height="40" fill="none" stroke="#334155" stroke-width="1.5"/>
                   <circle cx="50" cy="22" r="3" fill="#06b6d4" />
                   <circle cx="50" cy="38" r="3" fill="#06b6d4" />
                   <path d="M60,22 L70,22 M60,38 L70,38 M65,22 L65,38" stroke="#06b6d4" stroke-width="0.5"/>
                </svg>
                <span>AVE 测算已关闭</span>
              </div>

              <div v-else class="ave-flow animate-in">
                <div class="dir-row">
                  <button class="btn-ave" @click="openDirectionSelector"><i class="ri-edit-2-line"></i> 绘制方向</button>
                  <div class="dir-info-line">{{ form.directionLine || '未绘制' }}</div>
                </div>
                <div class="marker-row">
                  <button class="btn-ave mini" @click="openMarkerSelector('A')">选取 A 点</button>
                  <div class="res-box">{{ form.markerA || '--, --' }}</div>
                  <button class="btn-ave mini" @click="openMarkerSelector('B')">选取 B 点</button>
                  <div class="res-box">{{ form.markerB || '--, --' }}</div>
                </div>
                <div class="calib-row">
                  <button class="btn-ave mini" @click="openCalibration">比例标定</button>
                  <div class="calib-unit"> </div>
                  <input type="number" class="calib-input" v-model.number="form.pixLength" placeholder="像素" />
                  <div class="calib-unit">px = </div>
                  <input type="number" class="calib-input" v-model.number="form.physLength" placeholder="实际距离(mm)" />
                  <div class="calib-unit">mm</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
  
  <footer class="settings-footer">
    <button class="btn-primary" @click="handleSubmit">
      <i class="ri-checkbox-circle-fill"></i> 提交项目
    </button>
    <button class="btn-secondary" @click="$emit('cancel')">取消</button>
  </footer>
</template>

<script setup>
import { reactive, ref, computed, onMounted, onUnmounted } from 'vue';
import { Events } from '@wailsio/runtime';
import { GetActiveConfig, SaveProjectConfig, SelectDirectory } from '../../bindings/changeme/backend/projectservice';
import { OpenROISelector, OpenDirectionSelector, OpenCalibrationSelector, SetResolutionRatio } from '../../bindings/changeme/backend/hikcameraservice';

const props = defineProps({
  isCameraConnected: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['submit', 'cancel']);
const isSpeedOverridden = ref(false);

const form = reactive({
  experimenter: 'pims',
  sampleNo: '001',
  testDate: new Date().toLocaleString(),
  sampleShape: 'dogbone',
  width: 10.0,
  thickness: 2.0,
  diameter: 5.0,
  sectionLength: 25.0,
  type: 'tension',
  speed: 0.025,
  stopCondition: 'manual',
  filePath: 'C:/Users/Tang/Data',
  fileName: 'Result_001',
  // DIC
  dicEnable: false,
  externalTrigger: false,
  triggerType: 'internal',
  triggerInterval: 100,
  pulseWidth: 50,
  dicFolder: '',
  dicFileName: 'IMG_',
  // AVE
  videoExtEnable: false,
  markerA: '',
  markerB: '',
  directionLine: '',
  pixLength: 400,
  physLength: 25,
  poissonEnable: false
});

const currentArea = computed(() => {
  if (form.sampleShape === 'dogbone') {
    return (form.width * form.thickness).toFixed(3);
  }
  return (Math.PI * Math.pow(form.diameter / 2, 2)).toFixed(3);
});

const autoCalcSpeed = () => {
  if (!isSpeedOverridden.value) {
    form.speed = parseFloat((form.sectionLength * 0.001).toFixed(4));
  }
};

const handleSpeedManualChange = () => {
  const calculated = parseFloat((form.sectionLength * 0.001).toFixed(4));
  isSpeedOverridden.value = (form.speed !== calculated);
};

const handleSelectDir = async (field) => {
  try {
    const path = await SelectDirectory();
    if (path) form[field] = path;
  } catch (err) {
    console.error("Path selection failed", err);
  }
};

let offROISelected
let offDirectionSelected
let offCalibrationSelected

const formatROI = (roi) => {
  if (!roi) return '--, --'
  return `X:${Math.round(roi.x)} Y:${Math.round(roi.y)} W:${Math.round(roi.width)} H:${Math.round(roi.height)}`
}

const openMarkerSelector = async (label) => {
  try {
    await OpenROISelector(label)
  } catch (err) {
    alert(err)
  }
}

const openDirectionSelector = async () => {
  try {
    await OpenDirectionSelector()
  } catch (err) {
    alert(err)
  }
}

const openCalibration = async () => {
  try {
    await OpenCalibrationSelector()
  } catch (err) {
    alert(err)
  }
}

const handleROISelected = (payload) => {
  const data = payload?.data ?? payload
  const value = formatROI(data?.roi)
  if (data?.label === 'A') {
    form.markerA = value
  } else if (data?.label === 'B') {
    form.markerB = value
  }
}

const formatLine = (line) => {
  if (!line) return '未绘制'
  return `(${Math.round(line.x1)},${Math.round(line.y1)})→(${Math.round(line.x2)},${Math.round(line.y2)})`
}

const handleDirectionSelected = (payload) => {
  const data = payload?.data ?? payload
  if (data?.line) {
    form.directionLine = formatLine(data.line)
  }
}

const handleCalibrationSelected = (payload) => {
  const data = payload?.data ?? payload
  if (data?.length) {
    form.pixLength = data.length
  }
}

const handleSubmit = async() => {
  try {
    await SaveProjectConfig(form);
    if (form.pixLength && form.physLength) {
      await SetResolutionRatio(form.pixLength, form.physLength)
    }
    emit('submit', form);
  } catch (err) {
    alert("项目保存失败: " + err);
  }
};

onMounted(async () => {
  offROISelected = Events.On('hik_roi_selected', handleROISelected)
  offDirectionSelected = Events.On('hik_direction_selected', handleDirectionSelected)
  offCalibrationSelected = Events.On('hik_calibration_selected', handleCalibrationSelected)
  try {
    const config = await GetActiveConfig();
    if (config) {
      Object.assign(form, config);
    }
  } catch (err) {
    console.error("Failed to load config:", err);
  }
});

onUnmounted(() => {
  offROISelected?.()
  offDirectionSelected?.()
  offCalibrationSelected?.()
})
</script>

<style scoped>
.project-settings-root, .project-settings-content {
  display: flex;
  flex-direction: column;
  background: #0f172a;
  border-radius: 8px;
  border: 1px solid #334155;
  color: #f1f5f9;
  height: 100%;
  overflow: hidden;
  font-size: 13px;
}

.settings-header { padding: 12px 25px; background: #1e293b; border-bottom: 1px solid #334155; }
.title-wrap { display: flex; align-items: center; gap: 10px; color: #3b82f6; }
.title-wrap h2 { font-size: 1.1rem; margin: 0; font-weight: 800; }

.settings-body {
  flex: 1;
  padding: 15px;
  display: flex;
  flex-direction: column;
  gap: 15px;
  overflow-y: auto;
  min-height: 0;
}

.scrollable { overflow-y: auto; }

.settings-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 12px 20px;
  background: #1e293b;
  border-top: 1px solid #334155;
  flex-shrink: 0;
}

.config-section {
  border: 1px solid #334155; border-radius: 8px; padding: 15px;
  background: #1e293b; position: relative;
}

.section-tag {
  position: absolute; top: -10px; left: 15px; background: #3b82f6;
  padding: 1px 10px; font-size: 11px; font-weight: 900; border-radius: 3px; color: white;
}

/* 准静态网格 - 改为三栏 */
.quasi-grid { display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 0; margin-top: 5px; }
.sub-pane { padding: 0 15px; display: flex; flex-direction: column; }
.pane-title { font-size: 12px; color: #3b82f6; font-weight: 800; margin-bottom: 10px; display: block; }

.sample-preview { 
  height: 160px; background: #0f172a; border-radius: 6px; margin-top: 10px;
  display: flex; align-items: center; justify-content: center;
}
.sample-svg { width: 85%; height: 85%; }

.dim-inputs { display: flex; flex-direction: column; gap: 8px; }
.input-unit-group { display: flex; align-items: center; justify-content: space-between; }
.input-unit-group span { font-size: 12px; color: #94a3b8; }
.input-unit-group input { width: 80px; text-align: right; }
.area-calc { margin-top: 10px; font-size: 15px; color: #64748b; text-align: right; }
.area-calc span { color: #10b981; font-weight: bold; font-family: monospace; }

/* 准静态底部保存行 */
.quasi-footer { display: flex; gap: 15px; }
.flex-2 { flex: 2; }

/* 切换按钮 */
.toggle-group { display: flex; background: #0f172a; padding: 3px; border-radius: 3px; }
.toggle-group button { 
  flex: 1; border: none; background: transparent; color: #64748b; 
  padding: 2px; font-size: 12px; cursor: pointer; border-radius: 4px; font-weight: bold;
}
.toggle-group button.active { background: #334155; color: #f1f5f9; }

/* 路径选择器 */
.path-picker { display: flex; gap: 6px; }
.path-picker input { flex: 1; background: #0f172a; font-size: 12px; padding: 6px 12px; }
.btn-picker { 
  background: #334155; border: none; color: #3b82f6; padding: 0 12px; 
  border-radius: 5px; cursor: pointer; font-size: 16px;
}
.btn-picker:hover { background: #3b82f6; color: white; }

/* DIC & AVE 布局 */
.split-layout { display: flex; gap: 20px; min-height: 130px; margin-top: 5px; }
.left-controls { width: 120px; display: flex; flex-direction: column; gap: 15px; padding-top: 5px; }
.right-display { flex: 1; background: rgba(0,0,0,0.15); border-radius: 8px; padding: 12px; display: flex; flex-direction: column; justify-content: center; }

.placeholder { text-align: center; color: #334155; display: flex; flex-direction: column; align-items: center; gap: 5px; }
.placeholder i { font-size: 40px; opacity: 0.2; }
.diag-svg { width: 80px; opacity: 0.4; }

/* AVE 流程 */
.ave-flow { display: flex; flex-direction: column; gap: 10px; }
.dir-row { display: flex; align-items: center; gap: 10px; }
.dir-info-line { flex: 1; font-size: 11px; color: #06b6d4; background: rgba(6,182,212,0.1); border: 1px solid rgba(6,182,212,0.2); border-radius: 4px; padding: 8px 10px; text-align: center; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; line-height: 1.2; }
.marker-row { display: flex; align-items: center; gap: 8px; }
.marker-row .btn-ave.mini { flex-shrink: 0; }
.res-box { flex: 1; background: #0f172a; border: 1px solid #334155; border-radius: 4px; padding: 4px 6px; font-size: 10px; text-align: center; color: #3b82f6; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.calib-row { display: flex; align-items: center; gap: 8px; }
.calib-input { width: 80px; padding: 6px 8px; border: 1px solid #334155; border-radius: 4px; background: #0f172a; color: #f1f5f9; font-size: 11px; text-align: center; }
.calib-input::placeholder { color: #64748b; }

/* 复选框美化 */
.check-item { display: flex; align-items: center; gap: 8px; cursor: pointer; font-weight: bold; position: relative; }
.poisson-check { margin-top: 5px; font-size: 13px; color: #8b5cf6 !important; }
.mark { width: 18px; height: 18px; border: 2px solid #334155; border-radius: 4px; transition: 0.2s; position: relative; flex-shrink: 0; }
input[type="checkbox"]:checked + .mark { background: #3b82f6; border-color: #3b82f6; }
input[type="checkbox"]:checked + .mark::after {
  content: '✓'; position: absolute; top: 50%; left: 50%; transform: translate(-50%, -50%);
  color: white; font-size: 11px; font-weight: bold;
}
input[type="checkbox"] { display: none; }

/* 按钮 */
.btn-ave { 
  background: #334155; border: 1px solid #475569; color: white; padding: 8px 16px; 
  border-radius: 6px; font-size: 12px; cursor: pointer; font-weight: bold; min-width: 100px; line-height: 1.2;
}
.btn-ave.mini { padding: 6px 14px; font-size: 11px; line-height: 1.2; min-width: 80px; }

/* 通用输入 */
.form-row { display: flex; gap: 15px; }
.form-item { display: flex; flex-direction: column; gap: 4px; }
.form-item label { font-size: 12px; color: #94a3b8; font-weight: 800; }
input, select { 
  background: #0f172a; border: 1px solid #334155; border-radius: 6px; 
  padding: 8px 12px; color: #f1f5f9; font-size: 13px; outline: none; transition: 0.2s;
}
.danger-text { color: #ef4444 !important; border-color: #ef4444 !important; }

.settings-footer { padding: 15px 30px; display: flex; justify-content: flex-start; gap: 15px; background: #1e293b; border-top: 1px solid #334155; }
.btn-secondary { 
  background: transparent; 
  border: 1px solid #334155; 
  color: #64748b; 
  padding: 5px 25px; 
  border-radius: 6px; 
  cursor: pointer; 
  width: 150px;
  height: 40px;
}


.btn-primary { 
  background: #3b82f6; 
  color: white; 
  border: none; 
  padding: 5px 25px; 
  border-radius: 6px; 
  font-weight: 900; 
  cursor: pointer; 
  width: 150px;
  height: 40px;
}

.animate-in { animation: slideUp 0.3s ease-out; }
@keyframes slideUp { from { opacity: 0; transform: translateY(8px); } to { opacity: 1; transform: translateY(0); } }

.border-r { border-right: 1px solid #334155; }
.border-t { border-top: 1px solid #334155; }
.mt-15 { margin-top: 15px; }
.mt-10 { margin-top: 10px; }
.pt-15 { padding-top: 15px; }
.pt-10 { padding-top: 10px; }
.text-cyan { color: #06b6d4 !important; }
.flex-1 { flex: 1; }
.dual-row { display: flex; gap: 15px; }
</style>