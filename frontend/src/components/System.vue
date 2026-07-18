<template>
  <div class="system-page">
    <div v-if="!embedded" class="system-header">
      <div class="title-wrap">
        <i class="ri-settings-3-line"></i>
        <div>
          <h1>系统设置</h1>
          <p class="subtitle">配置电机、称重、限位与温度模块参数</p>
        </div>
      </div>
    </div>

    <div class="system-body scrollable">
      <!-- 电机模块 -->
      <div class="config-section">
        <div class="section-tag">电机模块</div>
        <div class="form-grid">
          <div class="form-item">
            <label>电机 ID</label>
            <input v-model.number="form.motor_id" type="number" min="0" max="255" />
          </div>
          <div class="form-item">
            <label>电机分辨率 (pulse/mm)</label>
            <input v-model.number="form.motor_resolution" type="number" step="0.00001" />
          </div>
          <div class="form-item">
            <label>电机方向</label>
            <select v-model.number="form.motor_direction">
              <option :value="1">正向 (+1)</option>
              <option :value="-1">反向 (-1)</option>
            </select>
          </div>
        </div>
      </div>

      <!-- 称重模块 -->
      <div class="config-section">
        <div class="section-tag color-weigh">压力模块</div>
        <div class="form-grid">
          <div class="form-item">
            <label>压力模块 ID</label>
            <input v-model.number="form.weigh_id" type="number" min="0" max="255" />
          </div>
          <div class="form-item">
            <label>压力分辨率 (unit/N)</label>
            <input v-model.number="form.weigh_resolution" type="number" step="0.0001" />
          </div>
          <div class="form-item">
            <label>压力方向</label>
            <select v-model.number="form.weigh_direction">
              <option :value="1">正向 (+1)</option>
              <option :value="-1">反向 (-1)</option>
            </select>
          </div>
        </div>
      </div>

      <!-- 限位模块 -->
      <div class="config-section">
        <div class="section-tag color-limit">限位模块</div>
        <div class="form-grid">
          <div class="form-item">
            <label>限位模块 ID</label>
            <input v-model.number="form.limit_id" type="number" min="0" max="255" />
          </div>
          <div class="form-item">
            <label>限位方向</label>
            <select v-model.number="form.limit_direction">
              <option :value="1">正向(端口1)&反向(端口2)</option>
              <option :value="-1">正向(端口2)&反向(端口1)</option>

            </select>
          </div>
          <div class="form-item checkbox-item">
            <label class="check-item">
              <input type="checkbox" v-model="form.limit_enabled" />
              <span class="mark"></span>
              启用限位模块
            </label>
          </div>
        </div>
      </div>

      <!-- 温度模块 -->
      <div class="config-section">
        <div class="section-tag color-temp">温度模块</div>
        <div class="form-grid">
          <div class="form-item">
            <label>温度模块 ID</label>
            <input v-model.number="form.temp_id" type="number" min="0" max="255" />
          </div>
          <div class="form-item checkbox-item">
            <label class="check-item">
              <input type="checkbox" v-model="form.temp_enabled" />
              <span class="mark"></span>
              启用温度模块
            </label>
          </div>
        </div>
      </div>
    </div>

    <div class="system-footer">
      <button class="btn-secondary" @click="resetToDefaults">恢复默认</button>
      <button class="btn-primary" @click="handleSave" :disabled="saving">
        <i class="ri-save-3-line"></i> {{ saving ? '保存中...' : '保存设置' }}
      </button>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted } from 'vue';
import * as SystemService from '../../bindings/MINIMTSPRO/backend/systemservice';

const defaultConfig = {
  motor_id: 1,
  motor_resolution: 1312.47251,
  motor_direction: 1,
  weigh_id: 2,
  weigh_resolution: 1.0,
  weigh_direction: 1,
  limit_id: 4,
  limit_enabled: false,
  limit_direction: 2,
  temp_id: 5,
  temp_enabled: true
};

const props = defineProps({
  embedded: {
    type: Boolean,
    default: false
  }
});

const form = reactive({ ...defaultConfig });
const saving = ref(false);
const emit = defineEmits(['close', 'saved']);

const loadConfig = async () => {
  try {
    const config = await SystemService.GetConfigFromLocalFile();
    if (config) {
      Object.assign(form, config);
    }
  } catch (err) {
    console.error('加载系统配置失败:', err);
    // alert('加载系统配置失败: ' + err);
  }
};

const handleSave = async () => {
  saving.value = true;
  try {
    await SystemService.UpdateConfigToLocalFile(form);
    emit('saved');
    // alert('系统设置已保存');
  } catch (err) {
    console.error('保存系统配置失败:', err);
    // alert('保存失败: ' + err);
  } finally {
    saving.value = false;
  }
};

const resetToDefaults = () => {
  if (!confirm('确定要恢复默认设置吗？')) return;
  Object.assign(form, defaultConfig);
};

onMounted(() => {
  loadConfig();
});
</script>

<style scoped>
@import 'remixicon/fonts/remixicon.css';

.system-page {
  display: flex;
  flex-direction: column;
  background: #0f172a;
  color: #f1f5f9;
  height: 100%;
  overflow: hidden;
  font-family: 'Inter', 'Segoe UI', system-ui, sans-serif;
}

.system-header {
  padding: 20px 30px;
  background: #1e293b;
  border-bottom: 1px solid #334155;
  flex-shrink: 0;
}

.title-wrap {
  display: flex;
  align-items: center;
  gap: 14px;
}

.title-wrap i {
  font-size: 28px;
  color: #3b82f6;
}

.system-header h1 {
  margin: 0;
  font-size: 20px;
  font-weight: 800;
}

.subtitle {
  margin: 4px 0 0 0;
  font-size: 13px;
  color: #94a3b8;
}

.system-body {
  flex: 1;
  padding: 24px 30px;
  display: flex;
  flex-direction: column;
  gap: 20px;
  overflow-y: auto;
  min-height: 0;
}

.scrollable {
  overflow-y: auto;
}

.config-section {
  border: 1px solid #334155;
  border-radius: 10px;
  padding: 20px;
  background: #1e293b;
  position: relative;
}

.section-tag {
  position: absolute;
  top: -10px;
  left: 18px;
  background: #3b82f6;
  padding: 2px 12px;
  font-size: 11px;
  font-weight: 800;
  border-radius: 4px;
  color: white;
}

.section-tag.color-weigh { background: #f59e0b; }
.section-tag.color-limit { background: #ef4444; }
.section-tag.color-temp { background: #10b981; }

.form-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  margin-top: 5px;
}

.form-item {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.form-item label {
  font-size: 12px;
  color: #94a3b8;
  font-weight: 700;
}

.form-item.checkbox-item {
  justify-content: center;
}

input,
select {
  background: #0f172a;
  border: 1px solid #334155;
  border-radius: 8px;
  padding: 8px 12px;
  color: #f1f5f9;
  font-size: 13px;
  outline: none;
  transition: 0.2s;
  box-sizing: border-box;
  width: 100%;
}

input:focus,
select:focus {
  border-color: #3b82f6;
}

/* 复选框美化 */
.check-item {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  font-weight: 600;
  font-size: 13px;
  color: #f1f5f9;
  position: relative;
}

.mark {
  width: 20px;
  height: 20px;
  border: 2px solid #334155;
  border-radius: 5px;
  transition: 0.2s;
  position: relative;
  flex-shrink: 0;
}

input[type="checkbox"]:checked + .mark {
  background: #3b82f6;
  border-color: #3b82f6;
}

input[type="checkbox"]:checked + .mark::after {
  content: '✓';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  color: white;
  font-size: 12px;
  font-weight: bold;
}

input[type="checkbox"] {
  display: none;
}

.system-footer {
  display: flex;
  justify-content: flex-end;
  gap: 16px;
  padding: 16px 30px;
  background: #1e293b;
  border-top: 1px solid #334155;
  flex-shrink: 0;
}

.btn-secondary {
  background: transparent;
  border: 1px solid #334155;
  color: #94a3b8;
  padding: 10px 24px;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 600;
  transition: all 0.2s;
}

.btn-secondary:hover {
  background: #334155;
  color: #f1f5f9;
}

.btn-primary {
  background: #3b82f6;
  color: white;
  border: none;
  padding: 10px 24px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 700;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: all 0.2s;
}

.btn-primary:hover:not(:disabled) {
  background: #2563eb;
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
